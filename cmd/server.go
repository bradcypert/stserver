package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bradcypert/stserver/internal"
	"github.com/bradcypert/stserver/internal/auth"
	"github.com/bradcypert/stserver/internal/events"
	"github.com/bradcypert/stserver/internal/handlers"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Setup Postgres
	dsn := os.Getenv("postgres_dsn")
	if dsn == "" {
		dsn = "postgres://dev:devpass@localhost:5432/sovereign?sslmode=disable"
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Println("Failed to connect to DB:", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Handle SIGINT/SIGTERM
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	gameEngine := internal.NewGameEngine(logger, rdb, pool)

	// Setup auth service
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-key-change-in-production"
	}
	authService := auth.NewService(jwtSecret)

	// Start tick engine in background
	go gameEngine.StartTickEngine(ctx)

	// Simulate a scheduled build
	err = scheduleBuildComplete(1, "trade_office", 10*time.Second)
	if err != nil {
		fmt.Println("Schedule error:", err)
	}

	// Auth endpoints
	authHandler := handlers.NewAuthHandler(pool, authService)
	http.HandleFunc("POST /auth/signup", authHandler.Signup)
	http.HandleFunc("POST /auth/login", authHandler.Login)
	http.HandleFunc("POST /auth/verify-email", authHandler.VerifyEmail)

	// Protected endpoints
	playerHandler := handlers.NewPlayerHandler(pool)
	http.HandleFunc("POST /players", authService.RequireAuth(playerHandler.CreatePlayer))

	// Faction endpoints
	factionHandler := handlers.NewFactionHandler(pool)
	http.HandleFunc("GET /factions", factionHandler.GetAllFactions)
	http.HandleFunc("GET /factions/{id}", factionHandler.GetFaction)
	http.HandleFunc("POST /factions/join", authService.RequireAuth(factionHandler.JoinFaction))
	http.HandleFunc("GET /player/faction", authService.RequireAuth(factionHandler.GetPlayerFaction))

	// Island Management endpoints
	islandHandler := handlers.NewIslandHandler(pool)
	http.HandleFunc("GET /my-island", authService.RequireAuth(islandHandler.GetPlayerIsland))
	http.HandleFunc("POST /my-island/buildings", authService.RequireAuth(islandHandler.ConstructBuilding))
	http.HandleFunc("POST /buildings/{building_id}/upgrade", authService.RequireAuth(islandHandler.UpgradeBuilding))
	http.HandleFunc("GET /building-types", islandHandler.GetBuildingTypes)
	http.HandleFunc("GET /building-production", islandHandler.GetBuildingProduction)

	go func() {
		logger.Info("Server started on :4200")
		if err := http.ListenAndServe(":4200", nil); err != http.ErrServerClosed {
			logger.Error("ListenAndServe(): %s\n", err)
			panic(err)
		}
	}()

	// Wait for shutdown signal
	<-sigs
	fmt.Println("\nShutting down...")

	cancel() // cancel the context so tick engine can stop

	// Give some time for cleanup (optional)
	time.Sleep(3 * time.Second)
	fmt.Println("Goodbye.")
}

func scheduleBuildComplete(portID int32, buildingType string, delay time.Duration) error {
	event := events.GameEvent{
		EventType:    events.GameEventPortBuilding,
		PortID:       portID,
		BuildingType: buildingType,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal build event: %w", err)
	}

	score := float64(time.Now().Add(delay).Unix())

	return rdb.ZAdd(ctx, "game_events", redis.Z{
		Score:  score,
		Member: data,
	}).Err()
}
