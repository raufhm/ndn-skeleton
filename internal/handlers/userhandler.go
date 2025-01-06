package handlers

import (
	"encoding/json"
	"github.com/ndn/internal/services"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

type UpdateUserRequest struct {
	Name string `json:"name" example:"John Doe" validate:"required"`
}

type UserResponse struct {
	ID        int64  `json:"id" example:"1"`
	Email     string `json:"email" example:"user@example.com"`
	Name      string `json:"name" example:"John Doe"`
	IsAdmin   bool   `json:"is_admin" example:"false"`
	CreatedAt string `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt string `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get the profile of the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} UserResponse
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/profile [get]
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		IsAdmin:   user.IsAdmin,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update the profile of the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Param request body UpdateUserRequest true "Update request"
// @Success 200 {object} UserResponse
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/profile [put]
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		h.sendError(w, "Name is required", http.StatusBadRequest)
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), userID, req.Name)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		IsAdmin:   user.IsAdmin,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get user details by ID (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} UserResponse
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /admin/users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		h.sendError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUser(r.Context(), id)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	response := UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		IsAdmin:   user.IsAdmin,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListUsers godoc
// @Summary List all users
// @Description Get a list of all users (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} UserResponse
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /admin/users [get]
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.ListUsers(r.Context())
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]UserResponse, len(users))
	for i, user := range users {
		response[i] = UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			IsAdmin:   user.IsAdmin,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
