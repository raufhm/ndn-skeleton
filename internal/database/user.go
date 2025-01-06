package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ndn/internal/models"

	"github.com/uptrace/bun"
)

type UserDB struct {
	db *bun.DB
}

func NewUserDB(db *bun.DB) *UserDB {
	return &UserDB{
		db: db,
	}
}

func (d *UserDB) GetUser(ctx context.Context, id int64) (*models.User, error) {
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

func (d *UserDB) ListUsers(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	err := d.db.NewSelect().
		Model(&users).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return users, nil
}

func (d *UserDB) UpdateUser(ctx context.Context, user *models.User) error {
	_, err := d.db.NewUpdate().
		Model(user).
		WherePK().
		OmitZero().
		Exec(ctx)

	return err
}
