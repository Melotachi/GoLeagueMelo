package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
)

type Team struct{
	TeamName string
	TeamId int
	TeamStrength int
	GoalsFor int
	GoalsAgainst int
	Wins int
	Draws int
	Losses int
	Points int
	GoalsDifference int
}

type Match struct{
	MatchId int
	Week int
	HomeTeam *Team
	AwayTeam *Team
	HomeTeamScore int
	AwayTeamScore int
	Played bool
}

type LeagueTableEntry struct{
	TeamName string
	Played int
	Wins int
	Draws int
	Losses int
	GoalsFor int
	GoalsAgainst int
	GoalsDifference int
	Points int
	Position int
}

type League struct {
	Teams []*Team
	Matches []*Match
	CurrentWeek int
	LeagueTable []*LeagueTableEntry
}

// create 4 random Premier League teams
func createPremierLeagueTeams() []*Team {
	teams := []*Team{
		{TeamName: "Manchester United", TeamId: 1, TeamStrength: 80},
		{TeamName: "Liverpool", TeamId: 2, TeamStrength: 85},
		{TeamName: "Manchester City", TeamId: 3, TeamStrength: 90},
		{TeamName: "Chelsea", TeamId: 4, TeamStrength: 88},
	}
	return teams
}

// create all matches for the league (home and away for each team pair)
func createPremierLeagueMatches(teams []*Team) []*Match {
	matches := []*Match{}
	matchId := 1
	week := 1

	// Define fixtures manually to ensure each team plays once per week
	// Week 1: Team 0 vs Team 1, Team 2 vs Team 3
	// Week 2: Team 0 vs Team 2, Team 1 vs Team 3
	// Week 3: Team 0 vs Team 3, Team 1 vs Team 2
	// Then repeat with reversed home/away for second leg
	
	weekFixtures := [][][2]int{
		{{0, 1}, {2, 3}}, // Week 1
		{{0, 2}, {1, 3}}, // Week 2
		{{0, 3}, {1, 2}}, // Week 3
	}
	
	// First leg
	for _, fixtures := range weekFixtures {
		for _, fixture := range fixtures {
			match := &Match{
				MatchId:       matchId,
				Week:          week,
				HomeTeam:      teams[fixture[0]],
				AwayTeam:      teams[fixture[1]],
				HomeTeamScore: 0,
				AwayTeamScore: 0,
				Played:        false,
			}
			matches = append(matches, match)
			matchId++
		}
		week++
	}
	
	// Second leg (reversed home/away)
	for _, fixtures := range weekFixtures {
		for _, fixture := range fixtures {
			match := &Match{
				MatchId:       matchId,
				Week:          week,
				HomeTeam:      teams[fixture[1]], // Reversed
				AwayTeam:      teams[fixture[0]], // Reversed
				HomeTeamScore: 0,
				AwayTeamScore: 0,
				Played:        false,
			}
			matches = append(matches, match)
			matchId++
		}
		week++
	}

	return matches
}

// simulate a single match based on team strength
func simulateMatch(match *Match) {
	if match.Played {
		return
	}

	homeTeam := match.HomeTeam
	awayTeam := match.AwayTeam

	// Calculate team strength difference and home advantage
	homeStrength := float64(homeTeam.TeamStrength) + 5.0 // +5 home advantage
	awayStrength := float64(awayTeam.TeamStrength)
	
	// Calculate attack potential based on strength (0.5 to 4.5 goals expected)
	homeAttack := (homeStrength / 100.0) * 4.0 + 0.5
	awayAttack := (awayStrength / 100.0) * 4.0 + 0.5
	
	// Add some randomness but weighted by strength
	homeRandomFactor := rand.Float64() * 2.0 - 1.0 // -1 to +1
	awayRandomFactor := rand.Float64() * 2.0 - 1.0 // -1 to +1
	
	homeExpected := homeAttack + homeRandomFactor
	awayExpected := awayAttack + awayRandomFactor
	
	// Ensure minimum 0 goals
	if homeExpected < 0 {
		homeExpected = 0
	}
	if awayExpected < 0 {
		awayExpected = 0
	}
	
	// Convert to actual goals (Poisson-like distribution simulation)
	homeTeamScore := int(homeExpected + 0.5) // Round to nearest int
	awayTeamScore := int(awayExpected + 0.5)
	
	// Cap maximum goals at 6
	if homeTeamScore > 6 {
		homeTeamScore = 6
	}
	if awayTeamScore > 6 {
		awayTeamScore = 6
	}

	match.HomeTeamScore = homeTeamScore
	match.AwayTeamScore = awayTeamScore

	// Update team stats
	homeTeam.GoalsFor += homeTeamScore
	awayTeam.GoalsFor += awayTeamScore
	homeTeam.GoalsAgainst += awayTeamScore
	awayTeam.GoalsAgainst += homeTeamScore

	// Update points and match results
	if homeTeamScore > awayTeamScore {
		homeTeam.Wins++
		awayTeam.Losses++
		homeTeam.Points += 3
	} else if homeTeamScore < awayTeamScore {
		awayTeam.Wins++
		homeTeam.Losses++
		awayTeam.Points += 3
	} else {
		homeTeam.Draws++
		awayTeam.Draws++
		homeTeam.Points += 1
		awayTeam.Points += 1
	}

	homeTeam.GoalsDifference = homeTeam.GoalsFor - homeTeam.GoalsAgainst
	awayTeam.GoalsDifference = awayTeam.GoalsFor - awayTeam.GoalsAgainst

	match.Played = true
}

// update the league table after each match
func updateLeagueTable(league *League){
	// at each week, the league table is deleted and recreated
	league.LeagueTable = []*LeagueTableEntry{}
	
	// Use the current team data from memory (not database during simulation)
	for _, team := range league.Teams {
		leagueTableEntry := LeagueTableEntry{
			TeamName: team.TeamName,
			Played: team.Wins + team.Draws + team.Losses,
			Wins: team.Wins,
			Draws: team.Draws,
			Losses: team.Losses,
			GoalsFor: team.GoalsFor,
			GoalsAgainst: team.GoalsAgainst,
			GoalsDifference: team.GoalsFor - team.GoalsAgainst,
			Points: team.Points,
		}
		league.LeagueTable = append(league.LeagueTable, &leagueTableEntry)
	}
	
	// Sort by points (descending), then by goal difference (descending)
	sort.Slice(league.LeagueTable, func(i, j int) bool {
		if league.LeagueTable[i].Points == league.LeagueTable[j].Points {
			return league.LeagueTable[i].GoalsDifference > league.LeagueTable[j].GoalsDifference
		}
		return league.LeagueTable[i].Points > league.LeagueTable[j].Points
	})
	
	// Assign positions
	for i, entry := range league.LeagueTable {
		entry.Position = i + 1
	}
}

func weeklySimulator(league *League){
	league.CurrentWeek++
	for _, match := range league.Matches {
		if match.Week == league.CurrentWeek && !match.Played {
			simulateMatch(match)
		}
	}
	updateLeagueTable(league)
}

func playSeason(league *League){
	// Calculate total weeks from matches
	totalWeeks := 0
	for _, match := range league.Matches { // find the last week of the season
		if match.Week > totalWeeks {
			totalWeeks = match.Week
		}
	}
	
	fmt.Printf("╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                    FOOTBALL LEAGUE SIMULATION                ║\n")
	fmt.Printf("║                     Total Matches: %-2d                       ║\n", len(league.Matches))
	fmt.Printf("║                     Total Weeks: %-2d                         ║\n", totalWeeks)
	fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n\n")
	
	for week := 1; week <= totalWeeks; week++ {
		weeklySimulator(league)
		
		fmt.Printf("┌─────────────────────────────────────────────────────────────┐\n")
		fmt.Printf("│                       WEEK %-2d RESULTS                       │\n", week)
		fmt.Printf("├─────────────────────────────────────────────────────────────┤\n")
		for _, match := range league.Matches {
			if match.Week == week && match.Played {
				fmt.Printf("│ %-20s %d - %-d %-20s             │\n", 
					match.HomeTeam.TeamName, match.HomeTeamScore,
					match.AwayTeamScore, match.AwayTeam.TeamName)
			}
		}
		fmt.Printf("└─────────────────────────────────────────────────────────────┘\n\n")
		
		fmt.Printf("┌─────────────────────────────────────────────────────────────┐\n")
		fmt.Printf("│                  LEAGUE TABLE AFTER WEEK %-2d                 │\n", week)
		fmt.Printf("├─────────────────────────────────────────────────────────────┤\n")
		fmt.Printf("│ %-20s %3s %3s %3s %3s %3s %4s │\n", "Team", "PTS", "P", "W", "D", "L", "GD")
		fmt.Printf("├─────────────────────────────────────────────────────────────┤\n")
		for _, entry := range league.LeagueTable {
			fmt.Printf("│ %-20s %3d %3d %3d %3d %3d %4d               │\n",
				entry.TeamName, entry.Points, entry.Played,
				entry.Wins, entry.Draws, entry.Losses, entry.GoalsDifference)
		}
		fmt.Printf("└─────────────────────────────────────────────────────────────┘\n")
		
		// Show championship predictions from week 4 onwards
		if week >= 4 {
			predictions := predictChampionship(league)
			fmt.Printf("\n┌─────────────────────────────────────────────────────────────┐\n")
			fmt.Printf("│            CHAMPIONSHIP PREDICTIONS AFTER WEEK %-2d           │\n", week)
			fmt.Printf("├─────────────────────────────────────────────────────────────┤\n")
			
			// Sort teams by prediction percentage
			type teamPrediction struct {
				name       string
				percentage float64
			}
			var sortedPredictions []teamPrediction
			for name, percentage := range predictions {
				sortedPredictions = append(sortedPredictions, teamPrediction{name, percentage})
			}
			
			// Simple sort by percentage (descending)
			for i := 0; i < len(sortedPredictions)-1; i++ {
				for j := i + 1; j < len(sortedPredictions); j++ {
					if sortedPredictions[i].percentage < sortedPredictions[j].percentage {
						sortedPredictions[i], sortedPredictions[j] = sortedPredictions[j], sortedPredictions[i]
					}
				}
			}
			
			for _, pred := range sortedPredictions {
				fmt.Printf("│ %-20s                               %5.1f%%   │\n", pred.name, pred.percentage)
			}
			fmt.Printf("└─────────────────────────────────────────────────────────────┘\n")
		}
		
		fmt.Println()
	}
}

// predict the championship percentages for each team
func predictChampionship(league *League) map[string]float64 {
	predictions := make(map[string]float64) // map of team name to prediction percentage
	
	// Calculate remaining matches for each team
	remainingMatches := make(map[string]int)
	for _, team := range league.Teams {
		remainingMatches[team.TeamName] = 0
	}
	
	for _, match := range league.Matches {
		if !match.Played {
			remainingMatches[match.HomeTeam.TeamName]++
			remainingMatches[match.AwayTeam.TeamName]++
		}
	}
	
	// Calculate maximum possible points for each team
	maxPossiblePoints := make(map[string]int)
	for _, entry := range league.LeagueTable {
		maxPossiblePoints[entry.TeamName] = entry.Points + (remainingMatches[entry.TeamName] * 3)
	}
	
	// Simple prediction algorithm based on:
	// 1. Current points (40%)
	// 2. Team strength (30%) 
	// 3. Goal difference (20%)
	// 4. Recent form/momentum (10%)
	
	totalWeight := 0.0
	teamWeights := make(map[string]float64)
	
	for _, entry := range league.LeagueTable {
		// Find team strength
		var teamStrength float64 = 75 // default
		for _, team := range league.Teams {
			if team.TeamName == entry.TeamName {
				teamStrength = float64(team.TeamStrength)
				break
			}
		}
		
		// Calculate weighted score
		pointsWeight := float64(entry.Points) * 0.4
		strengthWeight := (teamStrength / 100.0) * 30.0
		gdWeight := math.Max(float64(entry.GoalsDifference) * 0.2, 0)
		formWeight := float64(entry.Wins) * 1.0 // recent form approximation
		
		weight := pointsWeight + strengthWeight + gdWeight + formWeight
		
		// Bonus for being in top position
		if entry.Position == 1 {
			weight *= 1.2
		} else if entry.Position == 2 {
			weight *= 1.1
		}
		
		teamWeights[entry.TeamName] = weight
		totalWeight += weight
	}
	
	// Convert to percentages
	for teamName, weight := range teamWeights {
		if totalWeight > 0 {
			predictions[teamName] = (weight / totalWeight) * 100
		} else {
			predictions[teamName] = 25.0 // equal chance if no data
		}
	}
	
	return predictions
}

func declareChampions(league *League){
	fmt.Printf("\n╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                        FINAL RESULTS                         ║\n")
	fmt.Printf("╠══════════════════════════════════════════════════════════════╣\n")
	
	for _, entry := range league.LeagueTable {
		var trophy string
		switch entry.Position {
		case 1:
			trophy = "🏆"
		case 2:
			trophy = "🥈"
		case 3:
			trophy = "🥉"
		default:
			trophy = "  "
		}
		
		fmt.Printf("║ %s %-2d. %-20s %3d pts (%dW-%dD-%dL, %+d GD) ║\n", 
			trophy, entry.Position, entry.TeamName, entry.Points,
			entry.Wins, entry.Draws, entry.Losses, entry.GoalsDifference)
	}
	
	fmt.Printf("╠══════════════════════════════════════════════════════════════╣\n")
	
	for _, entry := range league.LeagueTable {
		if entry.Position == 1 {
			fmt.Printf("║                                                              ║\n")
			fmt.Printf("║                    🎉 CONGRATULATIONS! 🎉                    ║\n")
			fmt.Printf("║                                                              ║\n")
			fmt.Printf("║              %-20s IS THE CHAMPION!           ║\n", entry.TeamName)
			fmt.Printf("║                                                              ║\n")
			break
		}
	}
	
	fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n")
}

func main(){
	// Check if HTTP server mode is requested
	if len(os.Args) > 1 && os.Args[1] == "server" {
		startHTTPServer()
		return
	}
	
	teams := createPremierLeagueTeams()
	league := &League{
		Teams: teams,
		Matches: createPremierLeagueMatches(teams),
		CurrentWeek: 0,
		LeagueTable: []*LeagueTableEntry{},
	}
	
	// Play week by week and show results
	playSeason(league)
	declareChampions(league)
}