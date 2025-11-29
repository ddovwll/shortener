package repositories

import (
	"context"
	"fmt"
	"shortener/src/internal/domain/visit"

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
				"INSERT INTO visits (id, link_id, created_at, user_agent, ip) VALUES (%s, %s, %s, %s, %s)",
				pq.QuoteLiteral(visit.ID.String()),
				pq.QuoteLiteral(visit.LinkID.String()),
				pq.QuoteLiteral(visit.CreatedAt.String()),
				pq.QuoteLiteral(visit.UserAgent),
				pq.QuoteLiteral(visit.IPAddress),
			)
			visitsChan <- query
		}
	}()
}
