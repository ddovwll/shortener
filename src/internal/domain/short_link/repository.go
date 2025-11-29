package shortlink

import "context"

type ShortLinkRepository interface {
	Create(ctx context.Context, shortUrl, originalUrl string) (*ShortLink, error)
	Get(ctx context.Context, shortUrl string) (*ShortLink, error)
}
