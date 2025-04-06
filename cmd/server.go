package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type GameEvent struct {
	Type         string `json:"type"`
	PortID       string `json:"port_id"`
	BuildingType string `json:"building_type"`
}

var rdb = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

func startTickEngine() {
	ticker := time.NewTicker(5 * time.Second) // adjust tick speed
	defer ticker.Stop()

	for range ticker.C {
		now := float64(time.Now().Unix())

		// Get all due events
		events, err := rdb.ZRangeByScore(ctx, "game_events", &redis.ZRangeBy{
			Min:    "0",
			Max:    fmt.Sprintf("%.0f", now),
			Offset: 0,
			Count:  50, // limit to prevent overload
		}).Result()

		if err != nil {
			fmt.Println("Tick error:", err)
			continue
		}

		for _, raw := range events {
			var event GameEvent
			if err := json.Unmarshal([]byte(raw), &event); err != nil {
				fmt.Println("Unmarshal error:", err)
				continue
			}

			handleEvent(event)

			// Remove from Redis after handling
			rdb.ZRem(ctx, "game_events", raw)
		}
	}
}
