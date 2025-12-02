package repositories

import (
	"context"
	"errors"
	shortlink "shortener/src/internal/domain/short_link"
	"shortener/src/pkg/logger"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
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
		CreatedAt:   time.Now().UTC(),
	}

	query := `INSERT INTO short_links (id, short_code, original_url, created_at) VALUES ($1, $2, $3, $4)`

	_, err := r.db.ExecWithRetry(ctx, r.retry,
		query,
		shortLink.ID,
		shortLink.ShortCode,
		shortLink.OriginalURL,
		shortLink.CreatedAt,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && string(pqErr.Code) == "23505" {
			logger.Debug("pizda")
			return nil, shortlink.ErrShortLinkAlreadyExists
		}

		return nil, err
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
