package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/bradcypert/stserver/internal/auth"
	"github.com/bradcypert/stserver/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FactionHandler struct {
	queries *db.Queries
}

func NewFactionHandler(pool *pgxpool.Pool) *FactionHandler {
	return &FactionHandler{
		queries: db.New(pool),
	}
}

type joinFactionRequest struct {
	FactionID int32 `json:"faction_id"`
}

type joinFactionResponse struct {
	Message     string `json:"message"`
	PlayerID    int32  `json:"player_id"`
	FactionID   int32  `json:"faction_id"`
	FactionName string `json:"faction_name"`
}

type playerWithFactionResponse struct {
	Player      db.Player `json:"player"`
	FactionName string    `json:"faction_name"`
}

func (h *FactionHandler) GetAllFactions(w http.ResponseWriter, r *http.Request) {
	factions, err := h.queries.GetAllFactions(r.Context())
	if err != nil {
		http.Error(w, "failed to get factions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(factions)
}

func (h *FactionHandler) GetFaction(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "faction ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid faction ID", http.StatusBadRequest)
		return
	}

	faction, err := h.queries.GetFactionByID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "faction not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(faction)
}

func (h *FactionHandler) JoinFaction(w http.ResponseWriter, r *http.Request) {
	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "user not authenticated", http.StatusUnauthorized)
		return
	}

	var req joinFactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Validate faction exists
	faction, err := h.queries.GetFactionByID(r.Context(), req.FactionID)
	if err != nil {
		http.Error(w, "faction not found", http.StatusNotFound)
		return
	}

	// Get user's player
	user, err := h.queries.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Find player by email (since players are linked to users by email)
	player, err := h.queries.GetPlayerByEmail(r.Context(), user.Email)
	if err != nil {
		http.Error(w, "player not found", http.StatusNotFound)
		return
	}

	// Update player's faction
	err = h.queries.UpdatePlayerFaction(r.Context(), db.UpdatePlayerFactionParams{
		Faction:  req.FactionID,
		PlayerID: player.ID,
	})
	if err != nil {
		http.Error(w, "failed to update faction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(joinFactionResponse{
		Message:     "Successfully joined faction",
		PlayerID:    player.ID,
		FactionID:   req.FactionID,
		FactionName: faction.Name,
	})
}

func (h *FactionHandler) GetPlayerFaction(w http.ResponseWriter, r *http.Request) {
	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "user not authenticated", http.StatusUnauthorized)
		return
	}

	// Get user's player
	user, err := h.queries.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Find player by email
	player, err := h.queries.GetPlayerByEmail(r.Context(), user.Email)
	if err != nil {
		http.Error(w, "player not found", http.StatusNotFound)
		return
	}

	// Get faction information
	faction, err := h.queries.GetFactionByID(r.Context(), player.Faction)
	if err != nil {
		http.Error(w, "faction not found", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(playerWithFactionResponse{
		Player:      player,
		FactionName: faction.Name,
	})
}