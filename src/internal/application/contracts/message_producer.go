package contracts

import "context"

type MessageProducer interface {
	Produce(ctx context.Context, key, value []byte) error
}
