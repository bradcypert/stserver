package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/bradcypert/stserver/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventsHandler struct {
	queries *db.Queries
}

func NewEventsHandler(pool *pgxpool.Pool) *EventsHandler {
	return &EventsHandler{
		queries: db.New(pool),
	}
}

type createEventRequest struct {
	Port         int32  `json:"port"`
	EventType    string `json:"event_type"`
	BuildingType string `json:"building_type"`
}

func (h *PlayerHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req createEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// We need to validate that you CAN build this building.
	building, err := h.queries.GetBuildingByPortAndType(r.Context(), db.GetBuildingByPortAndTypeParams{
		PortID: req.Port,
		Type:   req.BuildingType,
	})
	if err != nil {
		http.Error(w, "could not get building: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// If validation passes, queue up the build event, after subtracting the resource cost from the players.

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(building)
}
