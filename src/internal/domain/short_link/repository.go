package shortlink

import "context"

type ShortLinkRepository interface {
	Create(ctx context.Context, shortURL, originalURL string) (*ShortLink, error)
	Get(ctx context.Context, shortURL string) (*ShortLink, error)
}
