package visit

import (
	"context"
)

type VisitService interface {
	CreateBatch(ctx context.Context, visits []Visit)
	Register(ctx context.Context, visit Visit) error
}
