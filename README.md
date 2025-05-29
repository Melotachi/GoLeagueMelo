# Football League Simulator

This project is a Go application that simulates a 4-team football league. It can run both in console mode and as an HTTP API server.

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

- Realistic match simulation based on team strength
- Home advantage calculation
- Championship predictions from week 4 onwards
- JSON API responses
- SQLite database persistence
- Edit match results functionality
- Testable architecture using interfaces
- Gorilla Mux router usage

## Database Schema

The application uses SQLite with the following tables:

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
git clone <repository>
cd insider
go mod tidy
go build
./main server
```

### Docker Deployment

Create a `Dockerfile`:

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main", "server"]
```

Build and run:

```bash
docker build -t football-league .
docker run -p 8080:8080 football-league
```

### Cloud Deployment (Heroku)

Create a `Procfile`:

```
web: ./main server
```

Deploy:

```bash
heroku create your-app-name
git push heroku main
```

## Dependencies

- `github.com/gorilla/mux` - For HTTP routing
- `github.com/mattn/go-sqlite3` - SQLite database driver
- `github.com/lib/pq` - PostgreSQL driver (optional)
