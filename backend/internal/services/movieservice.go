package services

import (
	"context"
	"errors"

	"github.com/ndn/backend/internal/models"
	"github.com/uptrace/bun"
)

type MovieService struct {
	db *bun.DB
}

func NewMovieService(db *bun.DB) *MovieService {
	return &MovieService{db: db}
}

type MovieFilter struct {
	CategoryID *int64   `json:"category_id,omitempty"`
	Search     string   `json:"search,omitempty"`
	SortBy     string   `json:"sort_by,omitempty"`
	Categories []string `json:"categories,omitempty"`
	Year       *int     `json:"year,omitempty"`
	Page       int      `json:"page,omitempty"`
	PageSize   int      `json:"page_size,omitempty"`
}

func (s *MovieService) GetMovies(ctx context.Context, filter MovieFilter) ([]models.Movie, int, error) {
	query := s.db.NewSelect().Model((*models.Movie)(nil))

	if filter.Search != "" {
		query.Where("title ILIKE ? OR description ILIKE ?",
			"%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	if filter.CategoryID != nil {
		query.Join("JOIN movie_categories AS mc ON mc.movie_id = movie.id").
			Where("mc.category_id = ?", *filter.CategoryID)
	}

	if len(filter.Categories) > 0 {
		query.Where("categories && ?", bun.In(filter.Categories))
	}

	if filter.Year != nil {
		query.Where("release_year = ?", *filter.Year)
	}

	// Get total count
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	offset := (filter.Page - 1) * filter.PageSize

	// Apply sorting
	switch filter.SortBy {
	case "title_asc":
		query.Order("title ASC")
	case "title_desc":
		query.Order("title DESC")
	case "year_asc":
		query.Order("release_year ASC")
	case "year_desc":
		query.Order("release_year DESC")
	case "rating_desc":
		query.Order("rating DESC")
	default:
		query.Order("created_at DESC")
	}

	var movies []models.Movie
	err = query.
		Limit(filter.PageSize).
		Offset(offset).
		Scan(ctx, &movies)

	return movies, total, err
}

func (s *MovieService) GetMovie(ctx context.Context, id int64) (*models.Movie, error) {
	movie := new(models.Movie)
	err := s.db.NewSelect().
		Model(movie).
		Where("id = ?", id).
		Scan(ctx)
	return movie, err
}

func (s *MovieService) CreateMovie(ctx context.Context, movie *models.Movie) error {
	exists, err := s.db.NewSelect().
		Model((*models.Movie)(nil)).
		Where("title = ?", movie.Title).
		Exists(ctx)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("movie already exists")
	}

	_, err = s.db.NewInsert().Model(movie).Exec(ctx)
	return err
}

func (s *MovieService) UpdateMovie(ctx context.Context, movie *models.Movie) error {
	exists, err := s.db.NewSelect().
		Model((*models.Movie)(nil)).
		Where("title = ? AND id != ?", movie.Title, movie.ID).
		Exists(ctx)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("movie title already taken")
	}

	_, err = s.db.NewUpdate().
		Model(movie).
		WherePK().
		OmitZero().
		Exec(ctx)
	return err
}

func (s *MovieService) DeleteMovie(ctx context.Context, id int64) error {
	// Delete associated records first
	_, err := s.db.NewDelete().
		Model((*models.MovieCategory)(nil)).
		Where("movie_id = ?", id).
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = s.db.NewDelete().
		Model((*models.UserFavorite)(nil)).
		Where("movie_id = ?", id).
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = s.db.NewDelete().
		Model((*models.Movie)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (s *MovieService) GetRelatedMovies(ctx context.Context, movieID int64, limit int) ([]models.Movie, error) {
	// Get the categories of the current movie
	var movie models.Movie
	err := s.db.NewSelect().
		Model(&movie).
		Where("id = ?", movieID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	// Find movies with similar categories
	var movies []models.Movie
	err = s.db.NewSelect().
		Model(&movies).
		Where("id != ?", movieID).
		Where("categories && ?", bun.In(movie.Categories)).
		Order("rating DESC").
		Limit(limit).
		Scan(ctx)

	return movies, err
}

func (s *MovieService) GetTopRatedMovies(ctx context.Context, limit int) ([]models.Movie, error) {
	var movies []models.Movie
	err := s.db.NewSelect().
		Model(&movies).
		Order("rating DESC").
		Limit(limit).
		Scan(ctx)
	return movies, err
}

func (s *MovieService) GetRecentlyAddedMovies(ctx context.Context, limit int) ([]models.Movie, error) {
	var movies []models.Movie
	err := s.db.NewSelect().
		Model(&movies).
		Order("created_at DESC").
		Limit(limit).
		Scan(ctx)
	return movies, err
}
