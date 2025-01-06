package routes

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/ndn/backend/internal/handlers"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(
	authHandler *handlers.AuthHandler,
	movieHandler *handlers.MovieHandler,
	categoryHandler *handlers.CategoryHandler,
	userHandler *handlers.UserHandler,
) *chi.Mux {
	r := chi.NewRouter()

	// Basic middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			// Auth routes
			r.Post("/auth/register", authHandler.Register)
			r.Post("/auth/login", authHandler.Login)
			r.Post("/auth/refresh", authHandler.Refresh)

			// Movie routes
			r.Get("/movies", movieHandler.GetMovies)
			r.Get("/movies/{id}", movieHandler.GetMovie)
			r.Get("/movies/top-rated", movieHandler.GetTopRatedMovies)
			r.Get("/movies/recently-added", movieHandler.GetRecentlyAddedMovies)

			// Category routes
			r.Get("/categories", categoryHandler.GetCategories)
			r.Get("/categories/{id}", categoryHandler.GetCategory)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authHandler.AuthMiddleware)

			// User routes
			r.Route("/users", func(r chi.Router) {
				r.Get("/profile", userHandler.GetProfile)
				r.Put("/profile", userHandler.UpdateProfile)
			})

			// Admin routes
			r.Route("/admin", func(r chi.Router) {
				r.Use(authHandler.AdminMiddleware)

				// Movie management
				r.Route("/movies", func(r chi.Router) {
					r.Post("/", movieHandler.CreateMovie)
					r.Put("/{id}", movieHandler.UpdateMovie)
					r.Delete("/{id}", movieHandler.DeleteMovie)
				})

				// Category management
				r.Route("/categories", func(r chi.Router) {
					r.Post("/", categoryHandler.CreateCategory)
					r.Delete("/{id}", categoryHandler.DeleteCategory)
				})

				// User management
				r.Route("/users", func(r chi.Router) {
					r.Get("/", userHandler.ListUsers)
					r.Get("/{id}", userHandler.GetUser)
				})
			})
		})
	})

	return r
}
