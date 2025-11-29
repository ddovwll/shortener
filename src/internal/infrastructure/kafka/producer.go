package kafka

import (
	"context"

	"github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"
)

type Producer struct {
	producer *kafka.Producer
	rerty    retry.Strategy
}

func NewKafkaProducer(producer *kafka.Producer, retry retry.Strategy) *Producer {
	return &Producer{
		producer: producer,
		rerty:    retry,
	}
}

func (p *Producer) Produce(ctx context.Context, key, value []byte) error {
	return p.producer.SendWithRetry(ctx, p.rerty, key, value)
}
