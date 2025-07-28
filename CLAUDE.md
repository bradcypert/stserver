# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based game server called "stserver" for a trading/empire building game called "Sovereign Tides". The server manages players, ports, buildings, and game events through a time-based tick system. The game is pirate/naval themed in a time-setting similar to that of Pirates of the Caribbean. 

## Development Commands

### Database Operations
- **Run migrations**: `goose -dir ./db/migrations postgres "postgres://dev:devpass@localhost:5432/sovereign?sslmode=disable" up`
- **Generate typesafe Go code from SQL**: `cd db && sqlc generate`

### Infrastructure
- **Start services**: `docker-compose up` (starts PostgreSQL and Redis)
- **Run server**: `go run cmd/server.go`

## Architecture

### Core Components
- **Game Engine** (`internal/gameloop.go`): Tick-based system running every 5 seconds that processes scheduled events from Redis
- **Event System** (`internal/events/`): Type-safe game events (building construction, resource collection, ship construction)
- **Database Layer** (`internal/db/`): SQLC-generated type-safe PostgreSQL queries and models
- **HTTP Handlers** (`internal/handlers/`): REST API endpoints for player and event management

### Database
- **PostgreSQL**: Primary storage for players, ports, buildings, and resources
- **Redis**: Event queue using sorted sets for time-based event scheduling
- **SQLC**: Code generation for type-safe database operations
- **Goose**: Database migration tool

### Key Data Models
- **Player**: Email, display name, faction affiliation
- **Port**: Game locations where players can build
- **Building**: Structures built at ports with types and levels
- **Resources**: Game economy items (managed via recent migrations)

### Event Processing Flow
1. Events scheduled in Redis sorted set with Unix timestamp scores
2. Game engine tick (5s intervals) queries for due events
3. Events processed and removed from queue
4. Database updated with event results

### Configuration
- Database DSN via `postgres_dsn` environment variable (defaults to local dev setup)
- Redis connection hardcoded to `localhost:6379`
- HTTP server runs on port `:4200`

## Development Notes

- Use SQLC for all database operations - avoid raw SQL in Go code
- Game events must be JSON-serializable and include proper type information
- All scheduled events go through Redis sorted sets with Unix timestamp scoring
- Server gracefully handles SIGINT/SIGTERM for proper shutdown
