package models

type CreateShortLinkRequest struct {
	ShortURL    *string `json:"shortURL,omitempty" validate:"omitempty,max=6"`
	OriginalURL string  `json:"originalUrl" validate:"required,url"`
}

func (r CreateShortLinkRequest) ShortURLString() string {
	if r.ShortURL == nil {
		return ""
	}

	return *r.ShortURL
}
