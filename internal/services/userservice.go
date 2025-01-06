package services

import (
	"context"
	"fmt"
	"github.com/ndn/internal/database"
	"github.com/ndn/internal/models"
)

type UserService struct {
	db *database.UserDB
}

func NewUserService(db *database.UserDB) *UserService {
	return &UserService{
		db: db,
	}
}

func (s *UserService) GetUser(ctx context.Context, id int64) (*models.User, error) {
	user, err := s.db.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *UserService) ListUsers(ctx context.Context) ([]*models.User, error) {
	users, err := s.db.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id int64, name string) (*models.User, error) {
	user, err := s.db.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.Name = name
	if err := s.db.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}
