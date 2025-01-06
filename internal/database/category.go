package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ndn/internal/models"

	"github.com/uptrace/bun"
)

type CategoryDB struct {
	db *bun.DB
}

func NewCategoryDB(db *bun.DB) *CategoryDB {
	return &CategoryDB{
		db: db,
	}
}

func (d *CategoryDB) GetCategories(ctx context.Context) ([]*models.Category, error) {
	var categories []*models.Category
	err := d.db.NewSelect().
		Model(&categories).
		Order("name ASC").
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (d *CategoryDB) GetCategory(ctx context.Context, id int64) (*models.Category, error) {
	category := new(models.Category)
	err := d.db.NewSelect().
		Model(category).
		Where("id = ?", id).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, errors.New("category not found")
	}
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (d *CategoryDB) CategoryExists(ctx context.Context, name string) (bool, error) {
	exists, err := d.db.NewSelect().
		Model((*models.Category)(nil)).
		Where("name = ?", name).
		Exists(ctx)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (d *CategoryDB) CreateCategory(ctx context.Context, category *models.Category) error {
	_, err := d.db.NewInsert().
		Model(category).
		Exec(ctx)

	return err
}

func (d *CategoryDB) DeleteCategory(ctx context.Context, id int64) error {
	_, err := d.db.NewDelete().
		Model((*models.Category)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

func (d *CategoryDB) CategoryInUse(ctx context.Context, id int64) (bool, error) {
	exists, err := d.db.NewSelect().
		Model((*models.MovieCategory)(nil)).
		Where("category_id = ?", id).
		Exists(ctx)

	if err != nil {
		return false, err
	}

	return exists, nil
}
