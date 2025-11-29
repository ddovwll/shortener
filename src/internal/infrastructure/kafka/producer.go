package kafka

import (
	"context"

	wbfkafka "github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"
)

type Producer struct {
	producer *wbfkafka.Producer
	rerty    retry.Strategy
}

func NewKafkaProducer(producer *wbfkafka.Producer, retry retry.Strategy) *Producer {
	return &Producer{
		producer: producer,
		rerty:    retry,
	}
}

func (p *Producer) Produce(ctx context.Context, key, value []byte) error {
	return p.producer.SendWithRetry(ctx, p.rerty, key, value)
}
