package container

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/ndn/backend/internal/config"
	"github.com/ndn/backend/internal/database"
	"github.com/ndn/backend/internal/handlers"
	"github.com/ndn/backend/internal/logger"
	"github.com/ndn/backend/internal/services"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"time"
)

// BuildContainer sets up the dependency injection container
func BuildContainer() *dig.Container {
	container := dig.New()

	// Core dependencies
	provideCore(container)

	// Database layer
	provideDatabase(container)

	// Services layer
	provideServices(container)

	// Handlers layer
	provideHandlers(container)

	return container
}

func provideCore(container *dig.Container) {
	// Provide config
	must(container.Provide(func() (*config.Config, error) {
		return config.LoadConfig("config.yaml")
	}))

	// Provide logger
	must(container.Provide(func(cfg *config.Config) (*zap.Logger, error) {
		return logger.NewLogger(cfg)
	}))

	// Provide NewRelic
	must(container.Provide(func(cfg *config.Config) (*newrelic.Application, error) {
		if !cfg.NewRelic.Enabled {
			return nil, nil
		}
		return newrelic.NewApplication(
			newrelic.ConfigAppName(cfg.NewRelic.AppName),
			newrelic.ConfigLicense(cfg.NewRelic.LicenseKey),
		)
	}))
}

func provideDatabase(container *dig.Container) {
	// Provide PostgreSQL connection
	must(container.Provide(func(cfg *config.Config, logger *zap.Logger) (*sql.DB, error) {
		// Construct database URL
		dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Database,
			cfg.Database.SSLMode,
		)

		// Run migrations first
		if err := database.RunMigrations(dbURL); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %v", err)
		}

		// Open PostgreSQL connection
		sqldb, err := sql.Open("postgres", dbURL)
		if err != nil {
			return nil, fmt.Errorf("failed to open database connection: %v", err)
		}

		// Configure connection pool
		sqldb.SetMaxOpenConns(cfg.Database.MaxOpenConns)
		sqldb.SetMaxIdleConns(cfg.Database.MaxIdleConns)
		sqldb.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime))

		// Verify connection
		if err := sqldb.PingContext(context.Background()); err != nil {
			sqldb.Close()
			return nil, fmt.Errorf("failed to ping database: %v", err)
		}

		logger.Info("successfully connected to database")
		return sqldb, nil
	}))

	// Provide bun.DB instance
	must(container.Provide(func(sqldb *sql.DB, logger *zap.Logger) *bun.DB {
		// Create bun.DB instance with PostgreSQL dialect
		bundb := bun.NewDB(sqldb, pgdialect.New())
		return bundb
	}))

	// Provide specific database repositories
	must(container.Provide(database.NewAuthDB))
	must(container.Provide(database.NewCategoryDB))
	must(container.Provide(database.NewUserDB))

}

func provideServices(container *dig.Container) {
	// Auth service with JWT configuration
	must(container.Provide(func(
		authDB *database.AuthDB,
		cfg *config.Config,
		logger *zap.Logger,
	) *services.AuthService {
		return services.NewAuthService(authDB, cfg.JWT.Secret)
	}))

	// Category service
	must(container.Provide(func(
		categoryDB *database.CategoryDB,
		logger *zap.Logger,
	) *services.CategoryService {
		return services.NewCategoryService(categoryDB)
	}))

	// User service
	must(container.Provide(func(
		userDB *database.UserDB,
		logger *zap.Logger,
	) *services.UserService {
		return services.NewUserService(userDB)
	}))
}

func provideHandlers(container *dig.Container) {
	// Auth handler
	must(container.Provide(func(
		authService *services.AuthService,
		logger *zap.Logger,
	) *handlers.AuthHandler {
		return handlers.NewAuthHandler(authService)
	}))

	// Category handler
	must(container.Provide(func(
		categoryService *services.CategoryService,
		logger *zap.Logger,
	) *handlers.CategoryHandler {
		return handlers.NewCategoryHandler(categoryService)
	}))

	// Movie handler
	must(container.Provide(func(
		movieService *services.MovieService,
		logger *zap.Logger,
	) *handlers.MovieHandler {
		return handlers.NewMovieHandler(movieService)
	}))

	// User handler
	must(container.Provide(func(
		userService *services.UserService,
		logger *zap.Logger,
	) *handlers.UserHandler {
		return handlers.NewUserHandler(userService)
	}))
}

// must panics if err is not nil
func must(err error) {
	if err != nil {
		panic(err)
	}
}
