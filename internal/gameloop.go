package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/bradcypert/stserver/internal/events"
	"github.com/bradcypert/stserver/internal/island"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type GameEngine struct {
	logger        *slog.Logger
	redis         *redis.Client
	pool          *pgxpool.Pool
	islandService *island.Service
}

func NewGameEngine(logger *slog.Logger, redis *redis.Client, pool *pgxpool.Pool) GameEngine {
	return GameEngine{
		logger:        logger,
		redis:         redis,
		pool:          pool,
		islandService: island.NewService(pool),
	}
}

func (engine *GameEngine) StartTickEngine(ctx context.Context) {
	engine.logger.Debug("Starting Ticker Engine")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Tick engine stopping...")
			return
		case <-ticker.C:
			engine.processDueEvents(ctx)
			engine.processResourceGeneration(ctx)
			engine.processCompletedConstructions(ctx)
		}
	}
}

func (engine *GameEngine) processDueEvents(ctx context.Context) {
	engine.logger.Debug("Processing Due Events")
	now := float64(time.Now().Unix())

	e, err := engine.redis.ZRangeByScore(ctx, "game_events", &redis.ZRangeBy{
		Min:    "0",
		Max:    fmt.Sprintf("%.0f", now),
		Offset: 0,
		Count:  250,
	}).Result()

	engine.logger.Debug("Collected events", slog.Int("Count", len(e)))

	if err != nil {
		fmt.Println("Tick error:", err)
		return
	}

	for _, raw := range e {
		var event events.GameEvent
		if err := json.Unmarshal([]byte(raw), &event); err != nil {
			fmt.Println("Unmarshal error:", err)
			continue
		}

		engine.handleEvent(ctx, event)
		engine.redis.ZRem(ctx, "game_events", raw)
	}
}

func (engine *GameEngine) handleEvent(_ context.Context, event events.GameEvent) {
	if (event.EventType == events.GameEventPortBuilding) && (event.BuildingType != "") {
		events.HandleBuildEvent(context.Background(), event, engine.pool)
	}
	engine.logger.Debug("Handling Event!",
		slog.Int("Event Type", int(event.EventType)),
		slog.Int("Port ID", int(event.PortID)),
		slog.String("Building Type", event.BuildingType),
	)
}

func (engine *GameEngine) processResourceGeneration(ctx context.Context) {
	engine.logger.Debug("Processing Resource Generation")
	err := engine.islandService.ProcessResourceGeneration(ctx)
	if err != nil {
		engine.logger.Error("Error processing resource generation", slog.String("error", err.Error()))
	}
}

func (engine *GameEngine) processCompletedConstructions(ctx context.Context) {
	engine.logger.Debug("Processing Completed Constructions")
	err := engine.islandService.CompleteConstructions(ctx)
	if err != nil {
		engine.logger.Error("Error processing completed constructions", slog.String("error", err.Error()))
	}
}
