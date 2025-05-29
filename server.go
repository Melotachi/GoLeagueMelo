package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Global league instance for the HTTP server
var globalLeague *League
var storageService StorageService

// SimulatorService interface for testing and business logic access
type SimulatorService interface {
	GetLeagueTable() []*LeagueTableEntry
	SimulateNextWeek() error
	SimulateAllMatches() error
	GetMatches() []*Match
}

// LeagueSimulatorService implements SimulatorService
type LeagueSimulatorService struct {
	league *League
}

func NewLeagueSimulatorService(league *League) *LeagueSimulatorService {
	return &LeagueSimulatorService{league: league}
}

func (s *LeagueSimulatorService) GetLeagueTable() []*LeagueTableEntry {
	return s.league.LeagueTable
}

func (s *LeagueSimulatorService) SimulateNextWeek() error {
	// Find the next week to simulate
	nextWeek := s.league.CurrentWeek + 1
	hasMatches := false
	
	for _, match := range s.league.Matches {
		if match.Week == nextWeek && !match.Played {
			hasMatches = true
			break
		}
	}
	
	if !hasMatches {
		return fmt.Errorf("no more matches to simulate")
	}
	
	weeklySimulator(s.league)
	
	// Save updated data to database
	if storageService != nil {
		// Update current week
		if err := storageService.UpdateCurrentWeek(s.league.CurrentWeek); err != nil {
			return fmt.Errorf("failed to update current week: %v", err)
		}
		
		// Save match results and team updates
		for _, match := range s.league.Matches {
			if match.Week == s.league.CurrentWeek && match.Played {
				if err := storageService.SaveMatchResult(match); err != nil {
					return fmt.Errorf("failed to save match result: %v", err)
				}
			}
		}
		
		// Update team statistics
		for _, team := range s.league.Teams {
			if err := storageService.UpdateTeam(team); err != nil {
				return fmt.Errorf("failed to update team: %v", err)
			}
		}
	}
	
	return nil
}

func (s *LeagueSimulatorService) SimulateAllMatches() error {
	// Calculate total weeks from matches
	totalWeeks := 0
	for _, match := range s.league.Matches {
		if match.Week > totalWeeks {
			totalWeeks = match.Week
		}
	}
	
	// Simulate all remaining weeks
	for week := s.league.CurrentWeek + 1; week <= totalWeeks; week++ {
		weeklySimulator(s.league)
		
		// Save updated data to database after each week
		if storageService != nil {
			// Update current week
			if err := storageService.UpdateCurrentWeek(s.league.CurrentWeek); err != nil {
				return fmt.Errorf("failed to update current week: %v", err)
			}
			
			// Save match results for this week
			for _, match := range s.league.Matches {
				if match.Week == s.league.CurrentWeek && match.Played {
					if err := storageService.SaveMatchResult(match); err != nil {
						return fmt.Errorf("failed to save match result: %v", err)
					}
				}
			}
			
			// Update team statistics
			for _, team := range s.league.Teams {
				if err := storageService.UpdateTeam(team); err != nil {
					return fmt.Errorf("failed to update team: %v", err)
				}
			}
		}
	}
	
	return nil
}

func (s *LeagueSimulatorService) GetMatches() []*Match {
	return s.league.Matches
}

// HTTP Handlers

// GET /league/table - Returns current league table in JSON format
func getLeagueTableHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if err := json.NewEncoder(w).Encode(globalLeague.LeagueTable); err != nil {
		http.Error(w, "Error encoding league table", http.StatusInternalServerError)
		return
	}
}

// POST /league/next-week - Simulates next week and returns current table
func simulateNextWeekHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	service := NewLeagueSimulatorService(globalLeague)
	
	if err := service.SimulateNextWeek(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if err := json.NewEncoder(w).Encode(globalLeague.LeagueTable); err != nil {
		http.Error(w, "Error encoding league table", http.StatusInternalServerError)
		return
	}
}

// POST /league/play-all - Simulates all remaining matches and returns final table
func simulateAllMatchesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	service := NewLeagueSimulatorService(globalLeague)
	
	if err := service.SimulateAllMatches(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if err := json.NewEncoder(w).Encode(globalLeague.LeagueTable); err != nil {
		http.Error(w, "Error encoding league table", http.StatusInternalServerError)
		return
	}
}

// GET /league/matches?week=<hafta_no> - Returns matches for specific week or all matches
func getMatchesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	weekParam := r.URL.Query().Get("week")
	
	var matchesToReturn []*Match
	
	if weekParam != "" {
		week, err := strconv.Atoi(weekParam)
		if err != nil {
			http.Error(w, "Invalid week parameter", http.StatusBadRequest)
			return
		}
		
		for _, match := range globalLeague.Matches {
			if match.Week == week {
				matchesToReturn = append(matchesToReturn, match)
			}
		}
	} else {
		matchesToReturn = globalLeague.Matches
	}
	
	if err := json.NewEncoder(w).Encode(matchesToReturn); err != nil {
		http.Error(w, "Error encoding matches", http.StatusInternalServerError)
		return
	}
}

// GET /league/matches - Returns all matches and their results
func getAllMatchesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if err := json.NewEncoder(w).Encode(globalLeague.Matches); err != nil {
		http.Error(w, "Error encoding matches", http.StatusInternalServerError)
		return
	}
}

// PUT /league/matches/{id} - Edit match result
func editMatchResultHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	vars := mux.Vars(r)
	matchIdStr := vars["id"]
	
	matchId, err := strconv.Atoi(matchIdStr)
	if err != nil {
		http.Error(w, "Invalid match ID", http.StatusBadRequest)
		return
	}
	
	// Parse request body
	var requestBody struct {
		HomeScore int `json:"home_score"`
		AwayScore int `json:"away_score"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Find the match
	var targetMatch *Match
	for _, match := range globalLeague.Matches {
		if match.MatchId == matchId {
			targetMatch = match
			break
		}
	}
	
	if targetMatch == nil {
		http.Error(w, "Match not found", http.StatusNotFound)
		return
	}
	
	if !targetMatch.Played {
		http.Error(w, "Cannot edit unplayed match", http.StatusBadRequest)
		return
	}
	
	// Revert old match statistics
	homeTeam := targetMatch.HomeTeam
	awayTeam := targetMatch.AwayTeam
	
	// Revert goals
	homeTeam.GoalsFor -= targetMatch.HomeTeamScore
	awayTeam.GoalsFor -= targetMatch.AwayTeamScore
	homeTeam.GoalsAgainst -= targetMatch.AwayTeamScore
	awayTeam.GoalsAgainst -= targetMatch.HomeTeamScore
	
	// Revert points and match results
	if targetMatch.HomeTeamScore > targetMatch.AwayTeamScore {
		homeTeam.Wins--
		awayTeam.Losses--
		homeTeam.Points -= 3
	} else if targetMatch.HomeTeamScore < targetMatch.AwayTeamScore {
		awayTeam.Wins--
		homeTeam.Losses--
		awayTeam.Points -= 3
	} else {
		homeTeam.Draws--
		awayTeam.Draws--
		homeTeam.Points -= 1
		awayTeam.Points -= 1
	}
	
	// Apply new match result
	targetMatch.HomeTeamScore = requestBody.HomeScore
	targetMatch.AwayTeamScore = requestBody.AwayScore
	
	// Update goals
	homeTeam.GoalsFor += targetMatch.HomeTeamScore
	awayTeam.GoalsFor += targetMatch.AwayTeamScore
	homeTeam.GoalsAgainst += targetMatch.AwayTeamScore
	awayTeam.GoalsAgainst += targetMatch.HomeTeamScore
	
	// Update points and match results
	if targetMatch.HomeTeamScore > targetMatch.AwayTeamScore {
		homeTeam.Wins++
		awayTeam.Losses++
		homeTeam.Points += 3
	} else if targetMatch.HomeTeamScore < targetMatch.AwayTeamScore {
		awayTeam.Wins++
		homeTeam.Losses++
		awayTeam.Points += 3
	} else {
		homeTeam.Draws++
		awayTeam.Draws++
		homeTeam.Points += 1
		awayTeam.Points += 1
	}
	
	// Update goal differences
	homeTeam.GoalsDifference = homeTeam.GoalsFor - homeTeam.GoalsAgainst
	awayTeam.GoalsDifference = awayTeam.GoalsFor - awayTeam.GoalsAgainst
	
	// Update league table
	updateLeagueTable(globalLeague)
	
	// Save to database
	if storageService != nil {
		if err := storageService.SaveMatchResult(targetMatch); err != nil {
			http.Error(w, fmt.Sprintf("Failed to save match: %v", err), http.StatusInternalServerError)
			return
		}
		
		if err := storageService.UpdateTeam(homeTeam); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update home team: %v", err), http.StatusInternalServerError)
			return
		}
		
		if err := storageService.UpdateTeam(awayTeam); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update away team: %v", err), http.StatusInternalServerError)
			return
		}
	}
	
	// Return updated league table
	if err := json.NewEncoder(w).Encode(globalLeague.LeagueTable); err != nil {
		http.Error(w, "Error encoding league table", http.StatusInternalServerError)
		return
	}
}

// setupRoutes configures all HTTP routes using gorilla/mux
func setupRoutes() *mux.Router {
	r := mux.NewRouter()
	
	// API endpoints
	r.HandleFunc("/league/table", getLeagueTableHandler).Methods("GET")
	r.HandleFunc("/league/next-week", simulateNextWeekHandler).Methods("POST")
	r.HandleFunc("/league/play-all", simulateAllMatchesHandler).Methods("POST")
	r.HandleFunc("/league/matches", getMatchesHandler).Methods("GET")
	r.HandleFunc("/league/matches/{id}", editMatchResultHandler).Methods("PUT")
	
	return r
}

// initializeLeague creates and initializes the global league instance
func initializeLeague() {
	// Initialize storage service (SQLite by default)
	var err error
	storageService, err = NewSQLStorageService("sqlite3", "./league.db")
	if err != nil {
		log.Fatalf("Failed to initialize storage service: %v", err)
	}
	
	// Initialize database with teams and matches if needed
	if err := storageService.(*SQLStorageService).InitializeTeamsAndMatches(); err != nil {
		log.Fatalf("Failed to initialize database data: %v", err)
	}
	
	// Load data from database
	teams, err := storageService.GetTeams()
	if err != nil {
		log.Fatalf("Failed to load teams from database: %v", err)
	}
	
	matches, err := storageService.GetMatches()
	if err != nil {
		log.Fatalf("Failed to load matches from database: %v", err)
	}
	
	currentWeek, err := storageService.GetCurrentWeek()
	if err != nil {
		log.Fatalf("Failed to load current week from database: %v", err)
	}
	
	globalLeague = &League{
		Teams:       teams,
		Matches:     matches,
		CurrentWeek: currentWeek,
		LeagueTable: []*LeagueTableEntry{},
	}
	
	// Initialize the league table
	updateLeagueTable(globalLeague)
}

// startHTTPServer starts the HTTP server on the specified port
func startHTTPServer() {
	// Initialize the league
	initializeLeague()
	
	// Setup routes
	router := setupRoutes()
	
	// Start server
	fmt.Println("Starting HTTP server on :8080")
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /league/table           - Get current league table")
	fmt.Println("  POST /league/next-week       - Simulate next week")
	fmt.Println("  POST /league/play-all        - Simulate all remaining matches")
	fmt.Println("  GET  /league/matches         - Get all matches")
	fmt.Println("  GET  /league/matches?week=N  - Get matches for specific week")
	fmt.Println("  PUT  /league/matches/{id}    - Edit match result")
	
	log.Fatal(http.ListenAndServe(":8080", router))
} 