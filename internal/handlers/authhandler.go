package handlers

import (
	"encoding/json"
	"github.com/ndn/internal/services"
	"net/http"
	"strings"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com" validate:"required,email"`
	Password string `json:"password" example:"password123" validate:"required,min=8"`
}

type RegisterRequest struct {
	Email    string `json:"email" example:"user@example.com" validate:"required,email"`
	Password string `json:"password" example:"password123" validate:"required,min=8"`
	Name     string `json:"name" example:"John Doe" validate:"required"`
}

type AuthResponse struct {
	Token     string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresIn int64  `json:"expires_in" example:"3600"`
	UserID    int64  `json:"user_id" example:"1"`
	Name      string `json:"name" example:"John Doe"`
	Email     string `json:"email" example:"user@example.com"`
	IsAdmin   bool   `json:"is_admin" example:"false"`
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register request"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} ErrorResponse "Invalid request parameters"
// @Failure 409 {object} ErrorResponse "Email already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Email == "" || req.Password == "" || req.Name == "" {
		h.sendError(w, "Email, password, and name are required", http.StatusBadRequest)
		return
	}

	// Check if user exists
	exists, err := h.authService.UserExists(r.Context(), req.Email)
	if err != nil {
		h.sendError(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if exists {
		h.sendError(w, "Email already registered", http.StatusConflict)
		return
	}

	// Register user
	authResp, err := h.authService.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(authResp)
}

// Login godoc
// @Summary Login user
// @Description Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse "Invalid request parameters"
// @Failure 401 {object} ErrorResponse "Invalid credentials"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Email == "" || req.Password == "" {
		h.sendError(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Login user
	authResp, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			h.sendError(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}
		h.sendError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(authResp)
}

// Refresh godoc
// @Summary Refresh access token
// @Description Get a new access token using the refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} AuthResponse
// @Failure 401 {object} ErrorResponse "Invalid or expired token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	token := h.extractToken(r)
	if token == "" {
		h.sendError(w, "Missing authorization header", http.StatusUnauthorized)
		return
	}

	authResp, err := h.authService.RefreshToken(r.Context(), token)
	if err != nil {
		if err == services.ErrInvalidToken {
			h.sendError(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		h.sendError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(authResp)
}

// AuthMiddleware godoc
// @Summary Authentication middleware
// @Description Middleware to authenticate requests using JWT token
// @Security BearerAuth
func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := h.extractToken(r)
		if token == "" {
			h.sendError(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		userID, err := h.authService.ValidateToken(r.Context(), token)
		if err != nil {
			if err == services.ErrInvalidToken {
				h.sendError(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}
			h.sendError(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Add user ID to context
		ctx := services.ContextWithUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminMiddleware godoc
// @Summary Admin authorization middleware
// @Description Middleware to check if the authenticated user is an admin
// @Security BearerAuth
func (h *AuthHandler) AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := services.UserIDFromContext(r.Context())
		if userID == 0 {
			h.sendError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		isAdmin, err := h.authService.IsAdmin(r.Context(), userID)
		if err != nil {
			h.sendError(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !isAdmin {
			h.sendError(w, "Admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Helper functions

func (h *AuthHandler) extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if bearerToken == "" {
		return ""
	}

	parts := strings.Split(bearerToken, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

func (h *AuthHandler) sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
