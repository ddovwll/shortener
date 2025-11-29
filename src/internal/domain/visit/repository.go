package visit

import "context"

type VisitRepository interface {
	CreateBatch(ctx context.Context, visits []Visit)
}
