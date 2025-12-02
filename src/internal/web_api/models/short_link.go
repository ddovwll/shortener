package models

import (
	shortlink "shortener/src/internal/domain/short_link"
	"time"

	"github.com/google/uuid"
)

type CreateShortLinkRequest struct {
	ShortURL    *string `json:"shortURL,omitempty" validate:"omitempty,max=6"`
	OriginalURL string  `json:"originalURL" validate:"required,url"`
}

func (r CreateShortLinkRequest) ShortURLString() string {
	if r.ShortURL == nil {
		return ""
	}

	return *r.ShortURL
}

type CreateShortLinkResponse struct {
	ID          uuid.UUID `json:"id"`
	ShortCode   string    `json:"shortCode"`
	OriginalURL string    `json:"originalURL"`
	CreatedAt   time.Time `json:"createdAt"`
}

func ShortLinkToCreateResponse(shortLink shortlink.ShortLink) CreateShortLinkResponse {
	return CreateShortLinkResponse{
		ID:          shortLink.ID,
		ShortCode:   shortLink.ShortCode,
		OriginalURL: shortLink.OriginalURL,
		CreatedAt:   shortLink.CreatedAt,
	}
}
