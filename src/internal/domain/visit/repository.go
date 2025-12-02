package visit

import "context"

type VisitRepository interface {
	CreateBatch(ctx context.Context, visits []Visit)
	AnalyticsAggregatedByDay(ctx context.Context, shortURL string) ([]PeriodCount, error)
	AnalyticsAggregatedByMonth(ctx context.Context, shortURL string) ([]PeriodCount, error)
	AnalyticsAggregatedByUserAgent(ctx context.Context, shortURL string) ([]UserAgentCount, error)
}
