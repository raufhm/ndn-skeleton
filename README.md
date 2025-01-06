# NDN API Service

A modern Go-based API service for course/video streaming platform with clean architecture, observability, and API documentation.

## Architecture Overview

### Directory Structure
```
backend/
├── cmd/
│   └── server/          # Application entry point
├── docs/               # Auto-generated Swagger documentation
├── internal/           # Internal packages
│   ├── config/         # Configuration management
│   ├── container/      # Dependency injection
│   ├── database/       # Database connection and migrations
│   ├── handlers/       # HTTP request handlers
│   ├── logger/         # Logging configuration
│   ├── models/         # Data models
│   ├── newrelic/       # New Relic APM integration
│   ├── routes/         # Route definitions
│   └── services/       # Business logic
└── config.yaml         # Application configuration
```

### Key Components

#### 1. Configuration Management
- Uses YAML-based configuration with environment variable support
- Configuration structure defined in `config/config.go`
- Environment variables can be injected using `${VAR_NAME}` syntax
- Supports different environments (development, production)

#### 2. Dependency Injection
- Uses `uber-go/dig` for dependency injection
- Container configuration in `container/container.go`
- Provides clean initialization of services and handlers
- Supports testing through dependency mocking

#### 3. Database Layer
- Uses `uptrace/bun` as PostgreSQL ORM
- Models defined with struct tags for database mapping
- Supports migrations and schema versioning
- Connection pooling and configuration

#### 4. API Layer
- RESTful API using `go-chi/chi` router
- Middleware support for:
  - CORS
  - Authentication
  - Logging
  - Request tracing
- Swagger documentation for all endpoints

#### 5. Observability
- Structured logging using `uber-go/zap`
- New Relic APM integration for:
  - Request tracing
  - Performance monitoring
  - Error tracking
- Health check endpoints

## API Documentation

### Swagger Integration
The API is documented using Swagger/OpenAPI 3.0 specifications. Documentation is generated from code annotations using `swaggo/swag`.

#### Generating Documentation
```bash
# Install swag CLI
go install github.com/swaggo/swag/cmd/swag@latest

# Generate documentation
cd backend && swag init -g cmd/server/main.go
```

#### Accessing Documentation
- Swagger UI: `http://localhost:8080/swagger/index.html`
- Raw JSON: `http://localhost:8080/swagger/doc.json`
- Raw YAML: `http://localhost:8080/swagger/doc.yaml`

### Authentication
- JWT-based authentication
- Bearer token format
- Protected routes require `Authorization` header
- Token expiration and refresh mechanism

## Development Workflow

### 1. Setup
```bash
# Clone repository
git clone https://github.com/ndn.git
cd ndn

# Install dependencies
go mod download

# Copy and configure environment
cp backend/config.yaml.example backend/config.yaml
```

### 2. Running the Service
```bash
# Run the service
cd backend
go run cmd/server/main.go
```

### 3. Development Process
1. Update API handlers with Swagger annotations
2. Implement business logic in services
3. Add new routes in `routes/routes.go`
4. Update configuration if needed
5. Generate Swagger documentation
6. Test endpoints

### 4. Code Organization

#### Handlers
- Handle HTTP requests and responses
- Input validation
- Response formatting
- Swagger documentation annotations
- Error handling

Example handler structure:
```go
// @Summary Create a new movie
// @Description Create a new movie with the provided details
// @Tags movies
// @Accept json
// @Produce json
// @Param movie body CreateMovieRequest true "Movie details"
// @Success 201 {object} MovieResponse
// @Failure 400 {object} ErrorResponse
// @Router /movies [post]
func (h *Handler) CreateMovie(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

#### Services
- Business logic implementation
- Database operations
- External service integration
- Error handling and validation

#### Models
- Database schema definitions
- JSON serialization
- Validation rules
- Relationships

## Observability

### Logging
- JSON structured logging in production
- Development-friendly console logging in development
- Log levels: debug, info, warn, error
- Request logging middleware
- Error tracking

### APM with New Relic
- Request tracing
- Transaction monitoring
- Error tracking
- Performance metrics
- Distributed tracing

Configuration in `config.yaml`:
```yaml
newrelic:
  app_name: "NDN API"
  license_key: ${NEW_RELIC_LICENSE_KEY}
  enabled: true
  distributed_tracer_enabled: true

logger:
  level: debug
  encoding: json
```

## Security

### Authentication Flow
1. User registers or logs in
2. Server validates credentials
3. JWT token generated and returned
4. Token used for subsequent requests
5. Middleware validates token for protected routes

### Authorization
- Role-based access control
- Admin-only routes
- User-specific data access
- Middleware-based protection

## Error Handling
- Consistent error response format
- HTTP status code mapping
- Detailed error messages in development
- Sanitized errors in production
- Error logging and tracking

## Testing
- Unit tests for services
- Integration tests for handlers
- Mock interfaces for external dependencies
- Test coverage reporting

## Contributing
1. Fork the repository
2. Create a feature branch
3. Implement changes
4. Add tests
5. Update documentation
6. Submit pull request

## License
Apache 2.0 License 
