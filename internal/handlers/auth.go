package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/bradcypert/stserver/internal/auth"
	"github.com/bradcypert/stserver/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthHandler struct {
	queries     *db.Queries
	authService *auth.Service
}

func NewAuthHandler(pool *pgxpool.Pool, authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		queries:     db.New(pool),
		authService: authService,
	}
}

type signupRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Faction     int32  `json:"faction"`
}

type signupResponse struct {
	Message string `json:"message"`
	UserID  int32  `json:"user_id"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token   string    `json:"token"`
	UserID  int32     `json:"user_id"`
	Player  db.Player `json:"player,omitempty"`
	Message string    `json:"message,omitempty"`
}

type verifyEmailRequest struct {
	Token string `json:"token"`
}

type verifyEmailResponse struct {
	Message string `json:"message"`
	UserID  int32  `json:"user_id"`
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}
	if req.DisplayName == "" {
		http.Error(w, "display name is required", http.StatusBadRequest)
		return
	}
	if req.Faction == 0 {
		req.Faction = 1 // Default to "Unaffiliated"
	}

	// Validate password
	if err := h.authService.IsPasswordValid(req.Password); err != nil {
		http.Error(w, fmt.Sprintf("invalid password: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// Check if user already exists
	_, err := h.queries.GetUserByEmail(r.Context(), req.Email)
	if err == nil {
		http.Error(w, "user already exists with this email", http.StatusConflict)
		return
	}

	// Hash password
	passwordHash, err := h.authService.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "failed to process password", http.StatusInternalServerError)
		return
	}

	// Generate email verification token
	verificationToken, err := h.authService.GenerateEmailVerificationToken()
	if err != nil {
		http.Error(w, "failed to generate verification token", http.StatusInternalServerError)
		return
	}

	// Create user
	user, err := h.queries.CreateUser(r.Context(), db.CreateUserParams{
		Email:                      req.Email,
		PasswordHash:               passwordHash,
		EmailVerificationToken:     pgtype.Text{String: verificationToken, Valid: true},
		EmailVerificationExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
	})
	if err != nil {
		http.Error(w, "failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create player linked to user
	player, err := h.queries.CreatePlayer(r.Context(), db.CreatePlayerParams{
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Faction:     req.Faction,
	})
	if err != nil {
		http.Error(w, "failed to create player: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Link player to user
	err = h.queries.UpdatePlayerUserID(r.Context(), db.UpdatePlayerUserIDParams{
		UserID:   pgtype.Int4{Int32: user.ID, Valid: true},
		PlayerID: player.ID,
	})
	if err != nil {
		http.Error(w, "failed to link player to user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create player's starting island
	island, err := h.queries.CreatePlayerIsland(r.Context(), db.CreatePlayerIslandParams{
		PlayerID: player.ID,
		Name:     req.DisplayName + "'s Island",
		X:        rand.Int31n(1000), // Random position on game map
		Y:        rand.Int31n(1000),
	})
	if err != nil {
		http.Error(w, "failed to create starting island: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Initialize starting resources for the island
	err = h.queries.InitializePortResources(r.Context(), island.ID)
	if err != nil {
		http.Error(w, "failed to initialize island resources: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Add starting resources (wood, iron, gold, grain as defined in migration)
	err = h.queries.AddResourcesToPort(r.Context(), db.AddResourcesToPortParams{
		PortID: island.ID,
		Wood:   100,
		Iron:   20,
		Gold:   50,
		Grain:  25,
		// Other resources start at 0
	})
	if err != nil {
		http.Error(w, "failed to add starting resources: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Send verification email here
	// For now, we'll just return the token in the response (remove this in production)
	
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(signupResponse{
		Message: "User created successfully with starting island. Please check your email for verification instructions. (Dev: use token " + verificationToken + ")",
		UserID:  user.ID,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Get user by email
	user, err := h.queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check if email is verified
	if !user.EmailVerified {
		http.Error(w, "email not verified", http.StatusUnauthorized)
		return
	}

	// Verify password
	if err := h.authService.VerifyPassword(req.Password, user.PasswordHash); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := h.authService.GenerateJWT(user.ID)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	// Get associated player (if any)
	response := loginResponse{
		Token:  token,
		UserID: user.ID,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Get user by verification token
	user, err := h.queries.GetUserByVerificationToken(r.Context(), pgtype.Text{String: req.Token, Valid: true})
	if err != nil {
		http.Error(w, "invalid verification token", http.StatusBadRequest)
		return
	}

	// Check if token is expired
	if user.EmailVerificationExpiresAt.Valid && user.EmailVerificationExpiresAt.Time.Before(time.Now()) {
		http.Error(w, "verification token expired", http.StatusBadRequest)
		return
	}

	// Verify email
	err = h.queries.VerifyUserEmail(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "failed to verify email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(verifyEmailResponse{
		Message: "Email verified successfully",
		UserID:  user.ID,
	})
}