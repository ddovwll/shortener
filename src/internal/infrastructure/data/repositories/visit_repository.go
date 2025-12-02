package repositories

import (
	"context"
	"fmt"
	"shortener/src/internal/domain/visit"
	"shortener/src/pkg/logger"
	"time"

	"github.com/lib/pq"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type VisitRepository struct {
	db    *dbpg.DB
	retry retry.Strategy
}

func NewVisitRepository(db *dbpg.DB, retry retry.Strategy) *VisitRepository {
	return &VisitRepository{
		db:    db,
		retry: retry,
	}
}

const timeFormat = time.RFC3339

func (r *VisitRepository) CreateBatch(ctx context.Context, visits []visit.Visit) {
	visitsChan := make(chan string)
	r.db.BatchExec(ctx, visitsChan)
	go func() {
		defer close(visitsChan)
		for _, visit := range visits {
			if ctx.Err() != nil {
				return
			}
			query := fmt.Sprintf(
				`INSERT INTO visits (
                    id, link_id, created_at, user_agent, ip_address
                    ) VALUES (%s, %s, %s, %s, %s) ON CONFLICT DO NOTHING`,
				pq.QuoteLiteral(visit.ID.String()),
				pq.QuoteLiteral(visit.LinkID.String()),
				pq.QuoteLiteral(visit.CreatedAt.Format(timeFormat)),
				pq.QuoteLiteral(visit.UserAgent),
				pq.QuoteLiteral(visit.IPAddress),
			)
			visitsChan <- query
		}
	}()
}

func (r *VisitRepository) AnalyticsAggregatedByDay(ctx context.Context, shortURL string) ([]visit.PeriodCount, error) {
	query := `SELECT date_trunc('day', visits.created_at) as day, count(*) as count
				FROM visits
				JOIN public.short_links sl on sl.id = visits.link_id
				WHERE sl.short_code = $1
				GROUP BY day
				order by day
				`

	rows, err := r.db.QueryWithRetry(ctx, r.retry, query, shortURL)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			logger.Error("failed to close rows channel", "err", err)
		}
	}()

	var result []visit.PeriodCount
	for rows.Next() {
		var day time.Time
		var count int64
		if err := rows.Scan(&day, &count); err != nil {
			return nil, err
		}
		result = append(result, visit.PeriodCount{
			Period: day.Format("2006-01-02"),
			Count:  count,
		})
	}

	return result, nil
}

func (r *VisitRepository) AnalyticsAggregatedByMonth(
	ctx context.Context,
	shortURL string,
) ([]visit.PeriodCount, error) {
	query := `SELECT date_trunc('month', visits.created_at) as month, count(*) as count
				FROM visits
				JOIN public.short_links sl on sl.id = visits.link_id
				WHERE sl.short_code = $1
				GROUP BY month
				order by month
				`

	rows, err := r.db.QueryWithRetry(ctx, r.retry, query, shortURL)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			logger.Error("failed to close rows channel", "err", err)
		}
	}()

	var result []visit.PeriodCount
	for rows.Next() {
		var month time.Time
		var count int64
		if err := rows.Scan(&month, &count); err != nil {
			return nil, err
		}
		result = append(result, visit.PeriodCount{
			Period: month.Format("2006-01-02"),
			Count:  count,
		})
	}

	return result, nil
}

func (r *VisitRepository) AnalyticsAggregatedByUserAgent(
	ctx context.Context,
	shortURL string,
) ([]visit.UserAgentCount, error) {
	query := `SELECT user_agent, count(*) as count
				FROM visits
				JOIN public.short_links sl on sl.id = visits.link_id
				WHERE sl.short_code = $1
				GROUP BY user_agent
				`

	rows, err := r.db.QueryWithRetry(ctx, r.retry, query, shortURL)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			logger.Error("failed to close rows channel", "err", err)
		}
	}()

	var result []visit.UserAgentCount
	for rows.Next() {
		var userAgent string
		var count int64
		if err := rows.Scan(&userAgent, &count); err != nil {
			return nil, err
		}
		result = append(result, visit.UserAgentCount{
			UserAgent: userAgent,
			Count:     count,
		})
	}

	return result, nil
}
