package kafka

import (
	"context"
	"encoding/json"
	"shortener/src/internal/domain/visit"
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

func (c *VisitConsumer) Start(ctx context.Context) {
	msgs := make(chan kafka.Message)
	c.consumer.StartConsuming(ctx, msgs, c.retry)

	go c.run(ctx, msgs)
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
				// TODO: log
				if err := c.consumer.Commit(ctx, msg); err != nil {
					// TODO: log
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
		// TODO: log
		return
	}

	*batch = (*batch)[:0]
	*lastMsg = nil
}
