package shortlink

import (
	"time"

	"github.com/google/uuid"
)

const ShortLinkLength = 6

type ShortLink struct {
	ID          uuid.UUID
	ShortCode   string
	OriginalURL string
	CreatedAt   time.Time
}
