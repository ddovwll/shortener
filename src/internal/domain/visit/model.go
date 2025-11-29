package visit

import (
	"time"

	"github.com/google/uuid"
)

type Visit struct {
	ID        uuid.UUID
	LinkID    uuid.UUID
	CreatedAt time.Time
	UserAgent string
	IPAddress string
}
