package kafka

import (
	"context"
	"encoding/json"
	"shortener/src/internal/domain/visit"
	"shortener/src/pkg/logger"
	"time"

	"github.com/segmentio/kafka-go"
	wbfkafka "github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"
)

type VisitConsumer struct {
	consumer     *wbfkafka.Consumer
	visitService visit.VisitService
	retry        retry.Strategy
}

const (
	batchSize    = 100
	maxBatchWait = 5 * time.Second
)

func NewVisitConsumer(
	consumer *wbfkafka.Consumer,
	visitService visit.VisitService,
	retry retry.Strategy,
) *VisitConsumer {
	return &VisitConsumer{
		consumer:     consumer,
		visitService: visitService,
		retry:        retry,
	}
}

func (c *VisitConsumer) Start(ctx context.Context) {
	msgs := make(chan kafka.Message) // channel will close in StartConsuming
	c.consumer.StartConsuming(ctx, msgs, c.retry)

	go c.run(ctx, msgs)
}

func (c *VisitConsumer) Stop() error {
	return c.consumer.Close()
}

func (c *VisitConsumer) run(ctx context.Context, msgs <-chan kafka.Message) {
	batch := make([]visit.Visit, 0, batchSize)
	var lastMsg *kafka.Message

	timer := time.NewTicker(maxBatchWait)
	defer timer.Stop()

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				c.flush(context.WithoutCancel(ctx), &batch, &lastMsg)
				return
			}

			var v visit.Visit
			if err := json.Unmarshal(msg.Value, &v); err != nil {
				logger.Error("failed to unmarshal visit", err)
				if err := c.consumer.Commit(ctx, msg); err != nil {
					logger.Error("failed to commit visit", err)
				}
				continue
			}

			batch = append(batch, v)
			m := msg
			lastMsg = &m

			if len(batch) >= batchSize {
				c.flush(ctx, &batch, &lastMsg)
			}

		case <-timer.C:
			c.flush(ctx, &batch, &lastMsg)

		case <-ctx.Done():
			c.flush(context.WithoutCancel(ctx), &batch, &lastMsg)
			return
		}
	}
}

func (c *VisitConsumer) flush(ctx context.Context, batch *[]visit.Visit, lastMsg **kafka.Message) {
	if len(*batch) == 0 || *lastMsg == nil {
		return
	}

	c.visitService.CreateBatch(ctx, *batch)

	if err := c.consumer.Commit(ctx, **lastMsg); err != nil {
		logger.Error("failed to commit visit", err)
		return
	}

	*batch = (*batch)[:0]
	*lastMsg = nil
}
