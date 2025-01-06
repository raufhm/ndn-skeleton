package handlers

import (
	"encoding/json"
	"github.com/ndn/internal/models"
	"github.com/ndn/internal/services"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CategoryHandler struct {
	categoryService *services.CategoryService
}

func NewCategoryHandler(categoryService *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

type CreateCategoryRequest struct {
	Name string `json:"name" example:"Action"`
}

type CategoryResponse struct {
	ID   int64  `json:"id" example:"1"`
	Name string `json:"name" example:"Action"`
}

// GetCategories godoc
// @Summary Get all categories
// @Description Get a list of all movie categories
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {array} CategoryResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories [get]
func (h *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryService.GetCategories(r.Context())
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]CategoryResponse, len(categories))
	for i, category := range categories {
		response[i] = CategoryResponse{
			ID:   category.ID,
			Name: category.Name,
		}
	}

	json.NewEncoder(w).Encode(response)
}

// GetCategory godoc
// @Summary Get a category by ID
// @Description Get detailed information about a category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} CategoryResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories/{id} [get]
func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		h.sendError(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	category, err := h.categoryService.GetCategory(r.Context(), id)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	response := CategoryResponse{
		ID:   category.ID,
		Name: category.Name,
	}

	json.NewEncoder(w).Encode(response)
}

// CreateCategory godoc
// @Summary Create a new category
// @Description Create a new movie category
// @Tags categories
// @Accept json
// @Produce json
// @Param category body CreateCategoryRequest true "Category details"
// @Success 201 {object} CategoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /admin/categories [post]
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		h.sendError(w, "Category name is required", http.StatusBadRequest)
		return
	}

	category := &models.Category{
		Name: req.Name,
	}

	if err := h.categoryService.CreateCategory(r.Context(), category); err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := CategoryResponse{
		ID:   category.ID,
		Name: category.Name,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// DeleteCategory godoc
// @Summary Delete a category
// @Description Delete a category by ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /admin/categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		h.sendError(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	if err := h.categoryService.DeleteCategory(r.Context(), id); err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CategoryHandler) sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
