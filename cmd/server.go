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

	// Start tick engine in background
	go gameEngine.StartTickEngine(ctx)

	// Simulate a scheduled build
	err = scheduleBuildComplete(1, "trade_office", 10*time.Second)
	if err != nil {
		fmt.Println("Schedule error:", err)
	}

	playerHandler := handlers.NewPlayerHandler(pool)
	http.HandleFunc("POST /players", playerHandler.CreatePlayer)

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
	fmt.Println("Adding game event")
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
