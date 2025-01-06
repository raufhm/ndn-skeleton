# Swagger Documentation Guide

## Overview
This guide explains how to document and maintain the API using Swagger/OpenAPI 3.0 specifications with `swaggo/swag`.

## Setup

### 1. Dependencies
Required packages in `go.mod`:
```go
require (
    github.com/swaggo/swag v1.16.4
    github.com/swaggo/http-swagger/v2 v2.0.2
)
```

### 2. Installation
Install the Swagger CLI tool:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

## Documentation Structure

### 1. Main API Information
Located in `cmd/server/main.go`:
```go
// @title NDN API
// @version 1.0
// @description NDN API Service
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@ndn.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
```

### 2. Request/Response Models
Define models with examples:
```go
type CreateMovieRequest struct {
    Title       string   `json:"title" example:"The Matrix"`
    Description string   `json:"description" example:"A computer programmer..."`
    ReleaseYear int      `json:"release_year" example:"1999"`
    Duration    int      `json:"duration" example:"136"`
    Categories  []string `json:"categories" example:"['Action', 'Sci-Fi']"`
}

type ErrorResponse struct {
    Error string `json:"error" example:"Invalid request parameters"`
}
```

### 3. Endpoint Documentation
Document each endpoint with annotations:
```go
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
func (h *Handler) CreateMovie(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

## Common Annotations

### 1. Operation Metadata
- `@Summary`: Brief description
- `@Description`: Detailed description
- `@Tags`: Grouping tag
- `@Accept`: Request content type
- `@Produce`: Response content type
- `@Deprecated`: Mark as deprecated

### 2. Parameters
- `@Param`: Define parameters
  ```go
  // @Param name in type required description
  // @Param id path int true "Movie ID"
  // @Param limit query int false "Limit results"
  ```

### 3. Responses
- `@Success`: Successful response
- `@Failure`: Error response
- `@Header`: Response headers
  ```go
  // @Success 200 {object} MovieResponse
  // @Failure 404 {object} ErrorResponse
  ```

### 4. Security
- `@Security`: Required authentication
  ```go
  // @Security BearerAuth
  ```

## Generation and Serving

### 1. Generate Documentation
```bash
cd backend
swag init -g cmd/server/main.go
```

This creates:
- `docs/docs.go`
- `docs/swagger.json`
- `docs/swagger.yaml`

### 2. Serve Documentation
In `server/server.go`:
```go
router.Get("/swagger/*", httpSwagger.Handler(
    httpSwagger.URL("/swagger/doc.json"),
    httpSwagger.DeepLinking(true),
    httpSwagger.DocExpansion("none"),
    httpSwagger.DomID("swagger-ui"),
))
```

## Best Practices

### 1. Documentation
- Use clear, concise descriptions
- Include meaningful examples
- Document all possible responses
- Keep security requirements clear
- Use consistent naming

### 2. Organization
- Group related endpoints with tags
- Use consistent parameter naming
- Document all models with examples
- Include validation rules

### 3. Maintenance
- Update documentation with code changes
- Regenerate docs before commits
- Review documentation in PRs
- Keep examples up to date

## Common Issues

### 1. Generation Issues
- Missing main annotations
- Invalid model references
- Syntax errors in annotations
- Wrong file paths

### 2. Runtime Issues
- Missing route registration
- Incorrect content types
- Security definition mismatches
- Invalid example formats

## Testing Documentation

### 1. Validation
- Ensure generated JSON is valid
- Check all links work
- Verify security settings
- Test example requests

### 2. UI Testing
- Test in Swagger UI
- Verify all operations
- Check authorization
- Test response examples 