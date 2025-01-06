package services

import (
	"context"
	"fmt"

	"github.com/ndn/backend/internal/database"
	"github.com/ndn/backend/internal/models"
)

type CategoryService struct {
	db *database.CategoryDB
}

func NewCategoryService(db *database.CategoryDB) *CategoryService {
	return &CategoryService{
		db: db,
	}
}

func (s *CategoryService) GetCategories(ctx context.Context) ([]*models.Category, error) {
	return s.db.GetCategories(ctx)
}

func (s *CategoryService) GetCategory(ctx context.Context, id int64) (*models.Category, error) {
	category, err := s.db.GetCategory(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	return category, nil
}

func (s *CategoryService) CreateCategory(ctx context.Context, category *models.Category) error {
	exists, err := s.db.CategoryExists(ctx, category.Name)
	if err != nil {
		return fmt.Errorf("failed to check category existence: %w", err)
	}
	if exists {
		return fmt.Errorf("category already exists")
	}

	if err := s.db.CreateCategory(ctx, category); err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}
	return nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id int64) error {
	// Check if category exists
	_, err := s.db.GetCategory(ctx, id)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	// Check if category is being used by movies
	inUse, err := s.db.CategoryInUse(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check category usage: %w", err)
	}
	if inUse {
		return fmt.Errorf("category is being used by movies")
	}

	if err := s.db.DeleteCategory(ctx, id); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	return nil
}
