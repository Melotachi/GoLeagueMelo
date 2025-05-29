package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// StorageService interface for SQL database operations
type StorageService interface {
	SaveMatchResult(match *Match) error
	GetMatches() ([]*Match, error)
	GetTeams() ([]*Team, error)
	UpdateTeam(team *Team) error
	InitializeDatabase() error
	GetCurrentWeek() (int, error)
	UpdateCurrentWeek(week int) error
}

// SQLStorageService implements StorageService for SQL databases
type SQLStorageService struct {
	db         *sql.DB
	driverName string
}

// NewSQLStorageService creates a new SQL storage service
func NewSQLStorageService(driverName, dataSourceName string) (*SQLStorageService, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	service := &SQLStorageService{
		db:         db,
		driverName: driverName,
	}

	if err := service.InitializeDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	return service, nil
}

// InitializeDatabase creates the required tables
func (s *SQLStorageService) InitializeDatabase() error {
	// Create teams table
	teamsSQL := `
	CREATE TABLE IF NOT EXISTS teams (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		strength INTEGER NOT NULL,
		goals_for INTEGER DEFAULT 0,
		goals_against INTEGER DEFAULT 0,
		wins INTEGER DEFAULT 0,
		draws INTEGER DEFAULT 0,
		losses INTEGER DEFAULT 0,
		points INTEGER DEFAULT 0,
		goals_difference INTEGER DEFAULT 0
	)`

	if _, err := s.db.Exec(teamsSQL); err != nil {
		return fmt.Errorf("failed to create teams table: %v", err)
	}

	// Create matches table
	matchesSQL := `
	CREATE TABLE IF NOT EXISTS matches (
		id INTEGER PRIMARY KEY,
		week INTEGER NOT NULL,
		home_team_id INTEGER NOT NULL,
		away_team_id INTEGER NOT NULL,
		home_score INTEGER DEFAULT 0,
		away_score INTEGER DEFAULT 0,
		played BOOLEAN DEFAULT FALSE,
		FOREIGN KEY (home_team_id) REFERENCES teams(id),
		FOREIGN KEY (away_team_id) REFERENCES teams(id)
	)`

	if _, err := s.db.Exec(matchesSQL); err != nil {
		return fmt.Errorf("failed to create matches table: %v", err)
	}

	// Create league_state table for current week tracking
	leagueStateSQL := `
	CREATE TABLE IF NOT EXISTS league_state (
		id INTEGER PRIMARY KEY DEFAULT 1,
		current_week INTEGER DEFAULT 0
	)`

	if _, err := s.db.Exec(leagueStateSQL); err != nil {
		return fmt.Errorf("failed to create league_state table: %v", err)
	}

	// Initialize league state if not exists
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM league_state").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check league_state: %v", err)
	}

	if count == 0 {
		_, err := s.db.Exec("INSERT INTO league_state (current_week) VALUES (0)")
		if err != nil {
			return fmt.Errorf("failed to initialize league_state: %v", err)
		}
	}

	return nil
}

// SaveMatchResult saves or updates a match result
func (s *SQLStorageService) SaveMatchResult(match *Match) error {
	query := `
	INSERT OR REPLACE INTO matches (id, week, home_team_id, away_team_id, home_score, away_score, played)
	VALUES (?, ?, ?, ?, ?, ?, ?)`

	if s.driverName == "postgres" {
		query = `
		INSERT INTO matches (id, week, home_team_id, away_team_id, home_score, away_score, played)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			week = EXCLUDED.week,
			home_team_id = EXCLUDED.home_team_id,
			away_team_id = EXCLUDED.away_team_id,
			home_score = EXCLUDED.home_score,
			away_score = EXCLUDED.away_score,
			played = EXCLUDED.played`
	}

	_, err := s.db.Exec(query, match.MatchId, match.Week, match.HomeTeam.TeamId, 
		match.AwayTeam.TeamId, match.HomeTeamScore, match.AwayTeamScore, match.Played)
	
	if err != nil {
		return fmt.Errorf("failed to save match result: %v", err)
	}

	return nil
}

// GetMatches retrieves all matches from database
func (s *SQLStorageService) GetMatches() ([]*Match, error) {
	query := `
	SELECT m.id, m.week, m.home_team_id, m.away_team_id, m.home_score, m.away_score, m.played,
		   ht.name as home_name, ht.strength as home_strength,
		   at.name as away_name, at.strength as away_strength
	FROM matches m
	JOIN teams ht ON m.home_team_id = ht.id
	JOIN teams at ON m.away_team_id = at.id
	ORDER BY m.week, m.id`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query matches: %v", err)
	}
	defer rows.Close()

	var matches []*Match
	teamCache := make(map[int]*Team)

	for rows.Next() {
		var match Match
		var homeTeamId, awayTeamId int
		var homeName, awayName string
		var homeStrength, awayStrength int

		err := rows.Scan(&match.MatchId, &match.Week, &homeTeamId, &awayTeamId,
			&match.HomeTeamScore, &match.AwayTeamScore, &match.Played,
			&homeName, &homeStrength, &awayName, &awayStrength)
		if err != nil {
			return nil, fmt.Errorf("failed to scan match: %v", err)
		}

		// Get or create home team
		if homeTeam, exists := teamCache[homeTeamId]; exists {
			match.HomeTeam = homeTeam
		} else {
			homeTeam := &Team{
				TeamId:       homeTeamId,
				TeamName:     homeName,
				TeamStrength: homeStrength,
			}
			teamCache[homeTeamId] = homeTeam
			match.HomeTeam = homeTeam
		}

		// Get or create away team
		if awayTeam, exists := teamCache[awayTeamId]; exists {
			match.AwayTeam = awayTeam
		} else {
			awayTeam := &Team{
				TeamId:       awayTeamId,
				TeamName:     awayName,
				TeamStrength: awayStrength,
			}
			teamCache[awayTeamId] = awayTeam
			match.AwayTeam = awayTeam
		}

		matches = append(matches, &match)
	}

	return matches, nil
}

// GetTeams retrieves all teams from database
func (s *SQLStorageService) GetTeams() ([]*Team, error) {
	query := `
	SELECT id, name, strength, goals_for, goals_against, wins, draws, losses, points, goals_difference
	FROM teams
	ORDER BY id`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query teams: %v", err)
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		var team Team
		err := rows.Scan(&team.TeamId, &team.TeamName, &team.TeamStrength,
			&team.GoalsFor, &team.GoalsAgainst, &team.Wins, &team.Draws,
			&team.Losses, &team.Points, &team.GoalsDifference)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team: %v", err)
		}
		teams = append(teams, &team)
	}

	return teams, nil
}

// UpdateTeam updates team statistics
func (s *SQLStorageService) UpdateTeam(team *Team) error {
	query := `
	INSERT OR REPLACE INTO teams (id, name, strength, goals_for, goals_against, wins, draws, losses, points, goals_difference)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	if s.driverName == "postgres" {
		query = `
		INSERT INTO teams (id, name, strength, goals_for, goals_against, wins, draws, losses, points, goals_difference)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			strength = EXCLUDED.strength,
			goals_for = EXCLUDED.goals_for,
			goals_against = EXCLUDED.goals_against,
			wins = EXCLUDED.wins,
			draws = EXCLUDED.draws,
			losses = EXCLUDED.losses,
			points = EXCLUDED.points,
			goals_difference = EXCLUDED.goals_difference`
	}

	_, err := s.db.Exec(query, team.TeamId, team.TeamName, team.TeamStrength,
		team.GoalsFor, team.GoalsAgainst, team.Wins, team.Draws,
		team.Losses, team.Points, team.GoalsDifference)

	if err != nil {
		return fmt.Errorf("failed to update team: %v", err)
	}

	return nil
}

// GetCurrentWeek retrieves current week from database
func (s *SQLStorageService) GetCurrentWeek() (int, error) {
	var currentWeek int
	err := s.db.QueryRow("SELECT current_week FROM league_state WHERE id = 1").Scan(&currentWeek)
	if err != nil {
		return 0, fmt.Errorf("failed to get current week: %v", err)
	}
	return currentWeek, nil
}

// UpdateCurrentWeek updates current week in database
func (s *SQLStorageService) UpdateCurrentWeek(week int) error {
	query := "UPDATE league_state SET current_week = ? WHERE id = 1"
	if s.driverName == "postgres" {
		query = "UPDATE league_state SET current_week = $1 WHERE id = 1"
	}

	_, err := s.db.Exec(query, week)
	if err != nil {
		return fmt.Errorf("failed to update current week: %v", err)
	}
	return nil
}

// Close closes the database connection
func (s *SQLStorageService) Close() error {
	return s.db.Close()
}

// InitializeTeamsAndMatches populates the database with initial data
func (s *SQLStorageService) InitializeTeamsAndMatches() error {
	// Check if teams already exist
	teams, err := s.GetTeams()
	if err != nil {
		return err
	}

	if len(teams) > 0 {
		log.Println("Teams already exist in database, skipping initialization")
		return nil
	}

	// Create initial teams
	initialTeams := createPremierLeagueTeams()
	for _, team := range initialTeams {
		if err := s.UpdateTeam(team); err != nil {
			return fmt.Errorf("failed to initialize team %s: %v", team.TeamName, err)
		}
	}

	// Create initial matches
	initialMatches := createPremierLeagueMatches(initialTeams)
	for _, match := range initialMatches {
		if err := s.SaveMatchResult(match); err != nil {
			return fmt.Errorf("failed to initialize match %d: %v", match.MatchId, err)
		}
	}

	log.Println("Database initialized with teams and matches")
	return nil
} 