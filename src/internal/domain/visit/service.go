package visit

import (
	"context"
)

type VisitService interface {
	CreateBatch(ctx context.Context, visits []Visit)
	Register(ctx context.Context, visit Visit) error
	ByDayAnalytics(ctx context.Context, shortURL string) ([]PeriodCount, error)
	ByMonthAnalytics(ctx context.Context, shortURL string) ([]PeriodCount, error)
	ByUserAgentAnalytics(ctx context.Context, shortURL string) ([]UserAgentCount, error)
}
