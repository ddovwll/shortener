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

type PeriodCount struct {
	Period string `json:"period"`
	Count  int64  `json:"count"`
}

type UserAgentCount struct {
	UserAgent string `json:"userAgent"`
	Count     int64  `json:"count"`
}
