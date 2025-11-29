package contracts

import "context"

type MessageProducer interface {
	Produce(ctx context.Context, value []byte) error
}
