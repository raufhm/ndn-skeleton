package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ndn/internal/models"

	"github.com/uptrace/bun"
)

type AuthDB struct {
	db *bun.DB
}

func NewAuthDB(db *bun.DB) *AuthDB {
	return &AuthDB{
		db: db,
	}
}

func (d *AuthDB) CreateUser(ctx context.Context, user *models.User) error {
	_, err := d.db.NewInsert().
		Model(user).
		Exec(ctx)

	return err
}

func (d *AuthDB) GetUser(ctx context.Context, id int64) (*models.User, error) {
	user := new(models.User)
	err := d.db.NewSelect().
		Model(user).
		Where("id = ?", id).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (d *AuthDB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := new(models.User)
	err := d.db.NewSelect().
		Model(user).
		Where("email = ?", email).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (d *AuthDB) UserExists(ctx context.Context, email string) (bool, error) {
	exists, err := d.db.NewSelect().
		Model((*models.User)(nil)).
		Where("email = ?", email).
		Exists(ctx)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (d *AuthDB) UpdateUser(ctx context.Context, user *models.User) error {
	_, err := d.db.NewUpdate().
		Model(user).
		WherePK().
		Exec(ctx)

	return err
}
