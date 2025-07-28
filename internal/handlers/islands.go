package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/bradcypert/stserver/internal/auth"
	"github.com/bradcypert/stserver/internal/db"
	"github.com/bradcypert/stserver/internal/island"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IslandHandler struct {
	queries       *db.Queries
	islandService *island.Service
}

func NewIslandHandler(pool *pgxpool.Pool) *IslandHandler {
	return &IslandHandler{
		queries:       db.New(pool),
		islandService: island.NewService(pool),
	}
}

type constructBuildingRequest struct {
	BuildingType string `json:"building_type"`
}

type upgradeBuildingRequest struct {
	BuildingID int32 `json:"building_id"`
}

func (h *IslandHandler) GetPlayerIsland(w http.ResponseWriter, r *http.Request) {
	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "user not authenticated", http.StatusUnauthorized)
		return
	}

	// Get player's island
	user, err := h.queries.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	player, err := h.queries.GetPlayerByEmail(r.Context(), user.Email)
	if err != nil {
		http.Error(w, "player not found", http.StatusNotFound)
		return
	}

	port, err := h.queries.GetPortByPlayerId(r.Context(), player.ID)
	if err != nil {
		http.Error(w, "island not found", http.StatusNotFound)
		return
	}

	// Get island overview
	overview, err := h.islandService.GetIslandOverview(r.Context(), port.ID)
	if err != nil {
		http.Error(w, "failed to get island overview: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(overview)
}

func (h *IslandHandler) GetIslandOverview(w http.ResponseWriter, r *http.Request) {
	portIDStr := r.PathValue("port_id")
	if portIDStr == "" {
		http.Error(w, "port ID is required", http.StatusBadRequest)
		return
	}

	portID, err := strconv.ParseInt(portIDStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid port ID", http.StatusBadRequest)
		return
	}

	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "user not authenticated", http.StatusUnauthorized)
		return
	}

	// Verify the port belongs to the authenticated user
	user, err := h.queries.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	player, err := h.queries.GetPlayerByEmail(r.Context(), user.Email)
	if err != nil {
		http.Error(w, "player not found", http.StatusNotFound)
		return
	}

	port, err := h.queries.GetPortById(r.Context(), int32(portID))
	if err != nil {
		http.Error(w, "port not found", http.StatusNotFound)
		return
	}

	if port.PlayerID != player.ID {
		http.Error(w, "unauthorized access to port", http.StatusForbidden)
		return
	}

	// Get island overview
	overview, err := h.islandService.GetIslandOverview(r.Context(), int32(portID))
	if err != nil {
		http.Error(w, "failed to get island overview: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(overview)
}

func (h *IslandHandler) ConstructBuilding(w http.ResponseWriter, r *http.Request) {
	var req constructBuildingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "user not authenticated", http.StatusUnauthorized)
		return
	}

	// Get player's island
	user, err := h.queries.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	player, err := h.queries.GetPlayerByEmail(r.Context(), user.Email)
	if err != nil {
		http.Error(w, "player not found", http.StatusNotFound)
		return
	}

	port, err := h.queries.GetPortByPlayerId(r.Context(), player.ID)
	if err != nil {
		http.Error(w, "island not found", http.StatusNotFound)
		return
	}

	// Ensure port has resources table entry
	err = h.queries.InitializePortResources(r.Context(), port.ID)
	if err != nil {
		http.Error(w, "failed to initialize port resources: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Construct building
	building, err := h.islandService.ConstructBuilding(r.Context(), island.BuildingConstructionRequest{
		PortID:       port.ID,
		BuildingType: req.BuildingType,
	})
	if err != nil {
		http.Error(w, "failed to construct building: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(building)
}

func (h *IslandHandler) UpgradeBuilding(w http.ResponseWriter, r *http.Request) {
	buildingIDStr := r.PathValue("building_id")
	if buildingIDStr == "" {
		http.Error(w, "building ID is required", http.StatusBadRequest)
		return
	}

	buildingID, err := strconv.ParseInt(buildingIDStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid building ID", http.StatusBadRequest)
		return
	}

	// Get user ID from auth context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "user not authenticated", http.StatusUnauthorized)
		return
	}

	// Verify ownership
	user, err := h.queries.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	player, err := h.queries.GetPlayerByEmail(r.Context(), user.Email)
	if err != nil {
		http.Error(w, "player not found", http.StatusNotFound)
		return
	}

	building, err := h.queries.GetBuilding(r.Context(), int32(buildingID))
	if err != nil {
		http.Error(w, "building not found", http.StatusNotFound)
		return
	}

	port, err := h.queries.GetPortById(r.Context(), building.PortID)
	if err != nil {
		http.Error(w, "port not found", http.StatusNotFound)
		return
	}

	if port.PlayerID != player.ID {
		http.Error(w, "unauthorized access to building", http.StatusForbidden)
		return
	}

	// Upgrade building
	err = h.islandService.UpgradeBuilding(r.Context(), int32(buildingID))
	if err != nil {
		http.Error(w, "failed to upgrade building: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Building upgrade started"})
}

func (h *IslandHandler) GetBuildingTypes(w http.ResponseWriter, r *http.Request) {
	buildingTypes, err := h.queries.GetAllBuildingTypes(r.Context())
	if err != nil {
		http.Error(w, "failed to get building types: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(buildingTypes)
}

func (h *IslandHandler) GetBuildingProduction(w http.ResponseWriter, r *http.Request) {
	buildingType := r.URL.Query().Get("type")
	if buildingType == "" {
		http.Error(w, "building type is required", http.StatusBadRequest)
		return
	}

	production, err := h.queries.GetAllProductionForBuildingType(r.Context(), buildingType)
	if err != nil {
		http.Error(w, "failed to get production info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(production)
}