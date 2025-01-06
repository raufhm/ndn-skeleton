package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ndn/backend/internal/database"
	"github.com/ndn/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrUserNotFound       = errors.New("user not found")
)

type contextKey string

const (
	userIDKey contextKey = "user_id"
)

type AuthService struct {
	db        *database.AuthDB
	jwtSecret []byte
}

type Claims struct {
	UserID  int64  `json:"user_id"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

func NewAuthService(db *database.AuthDB, jwtSecret string) *AuthService {
	return &AuthService{
		db:        db,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, name string) (*AuthResponse, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Email:    email,
		Password: string(hashedPassword),
		Name:     name,
		IsAdmin:  false,
	}

	if err := s.db.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate token
	token, expiresIn, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{
		Token:     token,
		ExpiresIn: expiresIn,
		UserID:    user.ID,
		Name:      user.Name,
		Email:     user.Email,
		IsAdmin:   user.IsAdmin,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	// Get user by email
	user, err := s.db.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate token
	token, expiresIn, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{
		Token:     token,
		ExpiresIn: expiresIn,
		UserID:    user.ID,
		Name:      user.Name,
		Email:     user.Email,
		IsAdmin:   user.IsAdmin,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, token string) (*AuthResponse, error) {
	// Parse and validate token
	claims, err := s.parseToken(token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Get user
	user, err := s.db.GetUser(ctx, claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Generate new token
	newToken, expiresIn, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{
		Token:     newToken,
		ExpiresIn: expiresIn,
		UserID:    user.ID,
		Name:      user.Name,
		Email:     user.Email,
		IsAdmin:   user.IsAdmin,
	}, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (int64, error) {
	claims, err := s.parseToken(token)
	if err != nil {
		return 0, ErrInvalidToken
	}
	return claims.UserID, nil
}

func (s *AuthService) UserExists(ctx context.Context, email string) (bool, error) {
	return s.db.UserExists(ctx, email)
}

func (s *AuthService) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	user, err := s.db.GetUser(ctx, userID)
	if err != nil {
		return false, err
	}
	return user.IsAdmin, nil
}

// Helper functions

func (s *AuthService) generateToken(user *models.User) (string, int64, error) {
	// Token expiration time (24 hours)
	expirationTime := time.Now().Add(24 * time.Hour)
	expiresIn := int64(time.Until(expirationTime).Seconds())

	claims := &Claims{
		UserID:  user.ID,
		Email:   user.Email,
		IsAdmin: user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresIn, nil
}

func (s *AuthService) parseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// Context functions

func ContextWithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromContext(ctx context.Context) int64 {
	userID, _ := ctx.Value(userIDKey).(int64)
	return userID
}

// Response types

type AuthResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
	UserID    int64  `json:"user_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	IsAdmin   bool   `json:"is_admin"`
}
