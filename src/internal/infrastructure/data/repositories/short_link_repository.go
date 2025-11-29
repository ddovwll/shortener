package repositories

import (
	"context"
	shortlink "shortener/src/internal/domain/short_link"
	"shortener/src/pkg/logger"
	"time"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type ShortLinkRepository struct {
	db    *dbpg.DB
	retry retry.Strategy
}

func NewShortLinkRepository(db *dbpg.DB, retry retry.Strategy) *ShortLinkRepository {
	return &ShortLinkRepository{
		db:    db,
		retry: retry,
	}
}

func (r *ShortLinkRepository) Create(ctx context.Context, shortURL, originalURL string) (*shortlink.ShortLink, error) {
	shortLink := &shortlink.ShortLink{
		ID:          uuid.New(),
		ShortCode:   shortURL,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
	}

	query := `INSERT INTO short_links (id, short_code, original_url, created_at) VALUES ($1, $2, $3, $4)`

	res, err := r.db.ExecWithRetry(
		ctx,
		r.retry,
		query,
		shortLink.ID,
		shortLink.ShortCode,
		shortLink.OriginalURL,
		shortLink.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		logger.Error("failed to get number of rows affected", err)
		return shortLink, nil
	}

	if affected == 0 {
		return nil, shortlink.ErrShortLinkAlreadyExists
	}

	return shortLink, nil
}

func (r *ShortLinkRepository) Get(ctx context.Context, shortURL string) (*shortlink.ShortLink, error) {
	query := `SELECT * FROM short_links WHERE short_code = $1`

	row, err := r.db.QueryRowWithRetry(ctx, r.retry, query, shortURL)
	if err != nil {
		return nil, err
	}

	var shortLink shortlink.ShortLink

	err = row.Scan(&shortLink.ID, &shortLink.ShortCode, &shortLink.OriginalURL, &shortLink.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &shortLink, nil
}
