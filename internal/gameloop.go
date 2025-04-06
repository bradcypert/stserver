package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type GameEngine struct {
	logger *slog.Logger
	redis  *redis.Client
}

func NewGameEngine(logger *slog.Logger, redis *redis.Client) GameEngine {
	return GameEngine{
		logger,
		redis,
	}
}

type GameEventType int

const (
	GameEventPortBuilding GameEventType = iota
	GameEventResourceCollect
	GameEventShipConstruct
)

type GameEvent struct {
	PortID       string        `json:"port_id"`
	BuildingType string        `json:"building_type"`
	EventType    GameEventType `json:"event_type"`
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
			engine.logger.Debug("Tick")
			engine.processDueEvents(ctx)
		}
	}
}

func (engine *GameEngine) processDueEvents(ctx context.Context) {
	engine.logger.Debug("Processing Due Events")
	now := float64(time.Now().Unix())

	events, err := engine.redis.ZRangeByScore(ctx, "game_events", &redis.ZRangeBy{
		Min:    "0",
		Max:    fmt.Sprintf("%.0f", now),
		Offset: 0,
		Count:  250,
	}).Result()

	engine.logger.Debug("Collected events", slog.Int("Count", len(events)))

	if err != nil {
		fmt.Println("Tick error:", err)
		return
	}

	for _, raw := range events {
		var event GameEvent
		if err := json.Unmarshal([]byte(raw), &event); err != nil {
			fmt.Println("Unmarshal error:", err)
			continue
		}

		engine.handleEvent(ctx, event)
		engine.redis.ZRem(ctx, "game_events", raw)
	}
}

func (engine *GameEngine) handleEvent(_ context.Context, event GameEvent) {
	engine.logger.Debug("Handling Event!", slog.Int("Event Type", int(event.EventType)), event.PortID, event.BuildingType)
}
