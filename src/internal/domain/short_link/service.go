package shortlink

import "context"

type ShortLinkService interface {
	Create(ctx context.Context, shortURL, originalURL string) (*ShortLink, error)
	Get(ctx context.Context, shortURL string) (*ShortLink, error)
}
