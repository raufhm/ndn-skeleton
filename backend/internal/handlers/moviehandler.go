package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ndn/backend/internal/models"
	"github.com/ndn/backend/internal/services"
)

type MovieHandler struct {
	movieService *services.MovieService
}

func NewMovieHandler(movieService *services.MovieService) *MovieHandler {
	return &MovieHandler{
		movieService: movieService,
	}
}

type CreateMovieRequest struct {
	Title       string   `json:"title" example:"The Matrix"`
	Description string   `json:"description" example:"A computer programmer discovers a mysterious world..."`
	ReleaseYear int      `json:"release_year" example:"1999"`
	Duration    int      `json:"duration" example:"136"`
	PosterURL   string   `json:"poster_url" example:"https://example.com/matrix.jpg"`
	VideoURL    string   `json:"video_url" example:"https://example.com/matrix.mp4"`
	Categories  []string `json:"categories" example:"['Action', 'Sci-Fi']"`
}

type UpdateMovieRequest struct {
	Title       *string   `json:"title,omitempty" example:"The Matrix Reloaded"`
	Description *string   `json:"description,omitempty"`
	ReleaseYear *int      `json:"release_year,omitempty" example:"2003"`
	Duration    *int      `json:"duration,omitempty" example:"138"`
	PosterURL   *string   `json:"poster_url,omitempty"`
	VideoURL    *string   `json:"video_url,omitempty"`
	Categories  *[]string `json:"categories,omitempty"`
}

type MovieResponse struct {
	ID          int64    `json:"id" example:"1"`
	Title       string   `json:"title" example:"The Matrix"`
	Description string   `json:"description"`
	ReleaseYear int      `json:"release_year" example:"1999"`
	Duration    int      `json:"duration" example:"136"`
	PosterURL   string   `json:"poster_url"`
	VideoURL    string   `json:"video_url"`
	Categories  []string `json:"categories"`
	Rating      float64  `json:"rating" example:"4.8"`
}

type PaginatedMovieResponse struct {
	Movies []MovieResponse `json:"movies"`
	Total  int             `json:"total"`
	Page   int             `json:"page"`
}

// GetMovies godoc
// @Summary Get all movies
// @Description Get a paginated list of movies with optional filtering
// @Tags movies
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10)"
// @Param search query string false "Search term"
// @Param year query int false "Filter by year"
// @Param categories query []string false "Filter by categories"
// @Param sort_by query string false "Sort field (title, year, rating)"
// @Success 200 {object} PaginatedMovieResponse
// @Failure 500 {object} ErrorResponse
// @Router /movies [get]
func (h *MovieHandler) GetMovies(w http.ResponseWriter, r *http.Request) {
	filter := services.MovieFilter{
		Search:     r.URL.Query().Get("search"),
		SortBy:     r.URL.Query().Get("sort_by"),
		Categories: r.URL.Query()["categories"],
	}

	if yearStr := r.URL.Query().Get("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			filter.Year = &year
		}
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	} else {
		filter.Page = 1
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			filter.PageSize = pageSize
		}
	} else {
		filter.PageSize = 10
	}

	movies, total, err := h.movieService.GetMovies(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := PaginatedMovieResponse{
		Movies: make([]MovieResponse, len(movies)),
		Total:  total,
		Page:   filter.Page,
	}

	for i, movie := range movies {
		response.Movies[i] = MovieResponse{
			ID:          movie.ID,
			Title:       movie.Title,
			Description: movie.Description,
			ReleaseYear: movie.ReleaseYear,
			Duration:    movie.Duration,
			PosterURL:   movie.PosterURL,
			VideoURL:    movie.VideoURL,
			Categories:  movie.Categories,
			Rating:      movie.Rating,
		}
	}

	json.NewEncoder(w).Encode(response)
}

// GetMovie godoc
// @Summary Get a movie by ID
// @Description Get detailed information about a movie
// @Tags movies
// @Accept json
// @Produce json
// @Param id path int true "Movie ID"
// @Success 200 {object} MovieResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /movies/{id} [get]
func (h *MovieHandler) GetMovie(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	movie, err := h.movieService.GetMovie(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := MovieResponse{
		ID:          movie.ID,
		Title:       movie.Title,
		Description: movie.Description,
		ReleaseYear: movie.ReleaseYear,
		Duration:    movie.Duration,
		PosterURL:   movie.PosterURL,
		VideoURL:    movie.VideoURL,
		Categories:  movie.Categories,
		Rating:      movie.Rating,
	}

	json.NewEncoder(w).Encode(response)
}

// CreateMovie godoc
// @Summary Create a new movie
// @Description Create a new movie with the provided details
// @Tags movies
// @Accept json
// @Produce json
// @Param movie body CreateMovieRequest true "Movie details"
// @Success 201 {object} MovieResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /admin/movies [post]
func (h *MovieHandler) CreateMovie(w http.ResponseWriter, r *http.Request) {
	var req CreateMovieRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	movie := &models.Movie{
		Title:       req.Title,
		Description: req.Description,
		ReleaseYear: req.ReleaseYear,
		Duration:    req.Duration,
		PosterURL:   req.PosterURL,
		VideoURL:    req.VideoURL,
		Categories:  req.Categories,
	}

	if err := h.movieService.CreateMovie(r.Context(), movie); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := MovieResponse{
		ID:          movie.ID,
		Title:       movie.Title,
		Description: movie.Description,
		ReleaseYear: movie.ReleaseYear,
		Duration:    movie.Duration,
		PosterURL:   movie.PosterURL,
		VideoURL:    movie.VideoURL,
		Categories:  movie.Categories,
		Rating:      movie.Rating,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateMovie godoc
// @Summary Update a movie
// @Description Update an existing movie's details
// @Tags movies
// @Accept json
// @Produce json
// @Param id path int true "Movie ID"
// @Param movie body UpdateMovieRequest true "Movie details to update"
// @Success 200 {object} MovieResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /admin/movies/{id} [put]
func (h *MovieHandler) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	var req UpdateMovieRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	movie, err := h.movieService.GetMovie(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if req.Title != nil {
		movie.Title = *req.Title
	}
	if req.Description != nil {
		movie.Description = *req.Description
	}
	if req.ReleaseYear != nil {
		movie.ReleaseYear = *req.ReleaseYear
	}
	if req.Duration != nil {
		movie.Duration = *req.Duration
	}
	if req.PosterURL != nil {
		movie.PosterURL = *req.PosterURL
	}
	if req.VideoURL != nil {
		movie.VideoURL = *req.VideoURL
	}
	if req.Categories != nil {
		movie.Categories = *req.Categories
	}

	if err := h.movieService.UpdateMovie(r.Context(), movie); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := MovieResponse{
		ID:          movie.ID,
		Title:       movie.Title,
		Description: movie.Description,
		ReleaseYear: movie.ReleaseYear,
		Duration:    movie.Duration,
		PosterURL:   movie.PosterURL,
		VideoURL:    movie.VideoURL,
		Categories:  movie.Categories,
		Rating:      movie.Rating,
	}

	json.NewEncoder(w).Encode(response)
}

// DeleteMovie godoc
// @Summary Delete a movie
// @Description Delete a movie by ID
// @Tags movies
// @Accept json
// @Produce json
// @Param id path int true "Movie ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /admin/movies/{id} [delete]
func (h *MovieHandler) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	if err := h.movieService.DeleteMovie(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetTopRatedMovies godoc
// @Summary Get top rated movies
// @Description Get a list of top rated movies
// @Tags movies
// @Accept json
// @Produce json
// @Param limit query int false "Number of movies to return (default: 10)"
// @Success 200 {array} MovieResponse
// @Failure 500 {object} ErrorResponse
// @Router /movies/top-rated [get]
func (h *MovieHandler) GetTopRatedMovies(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	movies, err := h.movieService.GetTopRatedMovies(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]MovieResponse, len(movies))
	for i, movie := range movies {
		response[i] = MovieResponse{
			ID:          movie.ID,
			Title:       movie.Title,
			Description: movie.Description,
			ReleaseYear: movie.ReleaseYear,
			Duration:    movie.Duration,
			PosterURL:   movie.PosterURL,
			VideoURL:    movie.VideoURL,
			Categories:  movie.Categories,
			Rating:      movie.Rating,
		}
	}

	json.NewEncoder(w).Encode(response)
}

// GetRecentlyAddedMovies godoc
// @Summary Get recently added movies
// @Description Get a list of recently added movies
// @Tags movies
// @Accept json
// @Produce json
// @Param limit query int false "Number of movies to return (default: 10)"
// @Success 200 {array} MovieResponse
// @Failure 500 {object} ErrorResponse
// @Router /movies/recently-added [get]
func (h *MovieHandler) GetRecentlyAddedMovies(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	movies, err := h.movieService.GetRecentlyAddedMovies(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]MovieResponse, len(movies))
	for i, movie := range movies {
		response[i] = MovieResponse{
			ID:          movie.ID,
			Title:       movie.Title,
			Description: movie.Description,
			ReleaseYear: movie.ReleaseYear,
			Duration:    movie.Duration,
			PosterURL:   movie.PosterURL,
			VideoURL:    movie.VideoURL,
			Categories:  movie.Categories,
			Rating:      movie.Rating,
		}
	}

	json.NewEncoder(w).Encode(response)
}
