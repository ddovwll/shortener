package visit

import (
	"context"
)

type VisitService interface {
	CreateBatch(ctx context.Context, visits []Visit)
}
