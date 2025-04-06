package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bradcypert/stserver/internal"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Handle SIGINT/SIGTERM
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	gameEngine := internal.NewGameEngine(logger, rdb)

	// Start tick engine in background
	go gameEngine.StartTickEngine(ctx)

	// Simulate a scheduled build
	err := scheduleBuildComplete("port-42", "trade_office", 10*time.Second)
	if err != nil {
		fmt.Println("Schedule error:", err)
	}

	// Wait for shutdown signal
	<-sigs
	fmt.Println("\nShutting down...")

	cancel() // cancel the context so tick engine can stop

	// Give some time for cleanup (optional)
	time.Sleep(1 * time.Second)
	fmt.Println("Goodbye.")
}

func scheduleBuildComplete(portID, buildingType string, delay time.Duration) error {
	fmt.Println("Adding game event")
	event := internal.GameEvent{
		EventType:    internal.GameEventPortBuilding,
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
