package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	Email     string    `bun:"email,unique,notnull" json:"email"`
	Password  string    `bun:"password,notnull" json:"-"`
	Name      string    `bun:"name,notnull" json:"name"`
	IsAdmin   bool      `bun:"is_admin,notnull,default:false" json:"is_admin"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`

	Profile *UserProfile `bun:"rel:has-one,join:id=user_id" json:"profile,omitempty"`
}

// BeforeAppend is called before the model is inserted/updated
func (u *User) BeforeAppend(ctx context.Context, query *bun.InsertQuery) error {
	u.UpdatedAt = time.Now()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	return nil
}

type UserProfile struct {
	bun.BaseModel `bun:"table:user_profiles,alias:up"`

	ID          int64     `bun:"id,pk,autoincrement" json:"id"`
	UserID      int64     `bun:"user_id,unique,notnull" json:"user_id"`
	Avatar      string    `bun:"avatar" json:"avatar"`
	Bio         string    `bun:"bio" json:"bio"`
	DateOfBirth time.Time `bun:"date_of_birth" json:"date_of_birth"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`

	User *User `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
}

// BeforeAppend is called before the model is inserted/updated
func (up *UserProfile) BeforeAppend(ctx context.Context, query *bun.InsertQuery) error {
	up.UpdatedAt = time.Now()
	if up.CreatedAt.IsZero() {
		up.CreatedAt = time.Now()
	}
	return nil
}

type Movie struct {
	bun.BaseModel `bun:"table:movies,alias:m"`

	ID          int64     `bun:"id,pk,autoincrement" json:"id"`
	Title       string    `bun:"title,notnull" json:"title"`
	Description string    `bun:"description,notnull" json:"description"`
	ReleaseYear int       `bun:"release_year,notnull" json:"release_year"`
	Duration    int       `bun:"duration,notnull" json:"duration"` // in minutes
	PosterURL   string    `bun:"poster_url,notnull" json:"poster_url"`
	VideoURL    string    `bun:"video_url,notnull" json:"video_url"`
	Categories  []string  `bun:"categories,array" json:"categories"`
	Rating      float64   `bun:"rating" json:"rating"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// BeforeAppend is called before the model is inserted/updated
func (m *Movie) BeforeAppend(ctx context.Context, query *bun.InsertQuery) error {
	m.UpdatedAt = time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	return nil
}

type UserFavorite struct {
	bun.BaseModel `bun:"table:user_favorites,alias:uf"`

	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	UserID    int64     `bun:"user_id,notnull" json:"user_id"`
	MovieID   int64     `bun:"movie_id,notnull" json:"movie_id"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`

	User  *User  `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	Movie *Movie `bun:"rel:belongs-to,join:movie_id=id" json:"movie,omitempty"`
}

type Category struct {
	bun.BaseModel `bun:"table:categories,alias:c"`

	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	Name      string    `bun:"name,notnull,unique" json:"name"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// BeforeAppend is called before the model is inserted/updated
func (c *Category) BeforeAppend(ctx context.Context, query *bun.InsertQuery) error {
	c.UpdatedAt = time.Now()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	return nil
}

type MovieCategory struct {
	bun.BaseModel `bun:"table:movie_categories,alias:mc"`

	MovieID    int64     `bun:"movie_id,pk" json:"movie_id"`
	CategoryID int64     `bun:"category_id,pk" json:"category_id"`
	CreatedAt  time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`

	Movie    *Movie    `bun:"rel:belongs-to,join:movie_id=id" json:"movie,omitempty"`
	Category *Category `bun:"rel:belongs-to,join:category_id=id" json:"category,omitempty"`
}
