package container

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/ndn/internal/config"
	database2 "github.com/ndn/internal/database"
	handlers2 "github.com/ndn/internal/handlers"
	"github.com/ndn/internal/logger"
	services2 "github.com/ndn/internal/services"
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
		if err := database2.RunMigrations(dbURL); err != nil {
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
	must(container.Provide(database2.NewAuthDB))
	must(container.Provide(database2.NewCategoryDB))
	must(container.Provide(database2.NewUserDB))

}

func provideServices(container *dig.Container) {
	// Auth service with JWT configuration
	must(container.Provide(func(
		authDB *database2.AuthDB,
		cfg *config.Config,
		logger *zap.Logger,
	) *services2.AuthService {
		return services2.NewAuthService(authDB, cfg.JWT.Secret)
	}))

	// Category service
	must(container.Provide(func(
		categoryDB *database2.CategoryDB,
		logger *zap.Logger,
	) *services2.CategoryService {
		return services2.NewCategoryService(categoryDB)
	}))

	// User service
	must(container.Provide(func(
		userDB *database2.UserDB,
		logger *zap.Logger,
	) *services2.UserService {
		return services2.NewUserService(userDB)
	}))
}

func provideHandlers(container *dig.Container) {
	// Auth handler
	must(container.Provide(func(
		authService *services2.AuthService,
		logger *zap.Logger,
	) *handlers2.AuthHandler {
		return handlers2.NewAuthHandler(authService)
	}))

	// Category handler
	must(container.Provide(func(
		categoryService *services2.CategoryService,
		logger *zap.Logger,
	) *handlers2.CategoryHandler {
		return handlers2.NewCategoryHandler(categoryService)
	}))

	// Movie handler
	must(container.Provide(func(
		movieService *services2.MovieService,
		logger *zap.Logger,
	) *handlers2.MovieHandler {
		return handlers2.NewMovieHandler(movieService)
	}))

	// User handler
	must(container.Provide(func(
		userService *services2.UserService,
		logger *zap.Logger,
	) *handlers2.UserHandler {
		return handlers2.NewUserHandler(userService)
	}))
}

// must panics if err is not nil
func must(err error) {
	if err != nil {
		panic(err)
	}
}
