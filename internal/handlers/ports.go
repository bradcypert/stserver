package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/bradcypert/stserver/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PortHandler struct {
	queries *db.Queries
}

func NewPortHandler(pool *pgxpool.Pool) *PortHandler {
	return &PortHandler{
		queries: db.New(pool),
	}
}

type createPortRequest struct {
	PlayerID int32  `json:"player_id"`
	Name     string `json:"name"`
	X        int32  `json:"x"`
	Y        int32  `json:"y"`
}

func (h *PortHandler) CreatePort(w http.ResponseWriter, r *http.Request) {
	var req createPortRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		http.Error(w, "port name is required", http.StatusBadRequest)
		return
	}

	port, err := h.queries.CreatePort(r.Context(), db.CreatePortParams{
		PlayerID: req.PlayerID,
		Name:     req.Name,
		X:        req.X,
		Y:        req.Y,
	})
	if err != nil {
		http.Error(w, "could not create port: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(port)
}

func (h *PortHandler) GetPort(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "port ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid port ID", http.StatusBadRequest)
		return
	}

	port, err := h.queries.GetPortById(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "could not get port: "+err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(port)
}