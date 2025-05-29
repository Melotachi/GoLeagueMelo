# Football League Simulator

This project is a Go application that simulates a 4-team football league with Premier League rules. It can run both in console mode and as an HTTP API server with SQLite database persistence.

## Installation

```bash
go mod tidy
go build
```

## Usage

### Console Mode

```bash
./main
```

### HTTP Server Mode

```bash
./main server
```

The server runs on port `:8080` by default.

## API Endpoints

### 1. GET /league/table

Returns the current league table in JSON format.

**Example:**

```bash
curl http://localhost:8080/league/table
```

**Response:**

```json
[
  {
    "TeamName": "Manchester United",
    "Played": 0,
    "Wins": 0,
    "Draws": 0,
    "Losses": 0,
    "GoalsFor": 0,
    "GoalsAgainst": 0,
    "GoalsDifference": 0,
    "Points": 0,
    "Position": 1
  }
]
```

### 2. POST /league/next-week

Simulates the next week and returns the current table.

**Example:**

```bash
curl -X POST http://localhost:8080/league/next-week
```

### 3. POST /league/play-all

Simulates all remaining matches and returns the final table.

**Example:**

```bash
curl -X POST http://localhost:8080/league/play-all
```

### 4. GET /league/matches

Returns all matches and their results.

**Example:**

```bash
curl http://localhost:8080/league/matches
```

### 5. GET /league/matches?week=N

Returns matches for a specific week.

**Example:**

```bash
curl "http://localhost:8080/league/matches?week=1"
```

### 6. PUT /league/matches/{id}

Edit the result of a played match and recalculate league table.

**Example:**

```bash
curl -X PUT http://localhost:8080/league/matches/1 \
  -H "Content-Type: application/json" \
  -d '{"home_score": 3, "away_score": 1}'
```

## Teams

- Manchester United (Strength: 80)
- Liverpool (Strength: 85)
- Manchester City (Strength: 90)
- Chelsea (Strength: 88)

## Features

- ✅ Realistic match simulation based on team strength and home advantage
- ✅ Championship predictions from week 4 onwards (intelligent algorithm)
- ✅ Complete REST API with JSON responses
- ✅ SQLite database persistence with automatic initialization
- ✅ Edit match results functionality with automatic recalculation
- ✅ Interface-based testable architecture
- ✅ Both console and server modes
- ✅ Docker containerization support
- ✅ Professional logging and error handling

## Database Schema

The application uses SQLite with automatic initialization. Tables:

### teams

```sql
CREATE TABLE teams (
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
);
```

### matches

```sql
CREATE TABLE matches (
    id INTEGER PRIMARY KEY,
    week INTEGER NOT NULL,
    home_team_id INTEGER NOT NULL,
    away_team_id INTEGER NOT NULL,
    home_score INTEGER DEFAULT 0,
    away_score INTEGER DEFAULT 0,
    played BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (home_team_id) REFERENCES teams(id),
    FOREIGN KEY (away_team_id) REFERENCES teams(id)
);
```

### league_state

```sql
CREATE TABLE league_state (
    id INTEGER PRIMARY KEY DEFAULT 1,
    current_week INTEGER DEFAULT 0
);
```

## Deployment

### Local Development

```bash
git clone https://github.com/Melotachi/GoLeagueMelo.git
cd GoLeagueMelo
go mod tidy
go build
./main server
```

### Docker Deployment

The application includes a multi-stage Dockerfile with CGO enabled for SQLite support.

```bash
# Build the image
docker build -t football-league .

# Run the container
docker run -p 8080:8080 football-league
```

**Docker Features:**

- Multi-stage build for smaller final image
- CGO enabled for SQLite compatibility
- Automatic database initialization
- Alpine Linux base for security and size
- Exposed on port 8080

### Testing the API

After starting the server, test the endpoints:

```bash
# Get initial league table
curl http://localhost:8080/league/table

# Simulate one week
curl -X POST http://localhost:8080/league/next-week

# Get matches for week 1
curl "http://localhost:8080/league/matches?week=1"

# Edit a match result (after playing some matches)
curl -X PUT http://localhost:8080/league/matches/1 \
  -H "Content-Type: application/json" \
  -d '{"home_score": 2, "away_score": 1}'

# Simulate all remaining matches
curl -X POST http://localhost:8080/league/play-all
```

## Architecture

The project follows clean architecture principles:

- **Interfaces**: `SimulatorService`, `StorageService` for testability
- **Separation of Concerns**: Business logic, HTTP handlers, and storage are separated
- **Dependency Injection**: Services are injected where needed
- **Error Handling**: Comprehensive error handling with proper HTTP status codes

## Dependencies

- `github.com/gorilla/mux` - HTTP routing and middleware
- `github.com/mattn/go-sqlite3` - SQLite database driver (requires CGO)
- `github.com/lib/pq` - PostgreSQL driver (optional, for future PostgreSQL support)
