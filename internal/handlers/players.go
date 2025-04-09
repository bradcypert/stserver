package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PlayerHandler struct {
	queries *db.Queries
}

func NewPlayerHandler(pool *pgxpool.Pool) *PlayerHandler {
	return &PlayerHandler{
		queries: db.New(pool),
	}
}

type createPlayerRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Faction     string `json:"faction"`
}

func (h *PlayerHandler) CreatePlayer(w http.ResponseWriter, r *http.Request) {
	var req createPlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	player, err := h.queries.CreatePlayer(r.Context(), db.CreatePlayerParams{
		Username:    req.Username,
		DisplayName: req.DisplayName,
		Faction:     req.Faction,
	})
	if err != nil {
		http.Error(w, "could not create player: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(player)
}
