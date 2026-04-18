package kafkaPkg

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/segmentio/kafka-go"
)

type Producer[T any] struct {
	writer *kafka.Writer
}

func NewProducer[T any](brokers []string, topic string) *Producer[T] {
	return &Producer[T]{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  topic,
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
		},
	}
}

func (p *Producer[T]) Send(ctx context.Context, event T) error {
	return p.sendWithRetry(ctx, event, 0)
}

func (p *Producer[T]) Retry(ctx context.Context, event T, currentRetry int) error {
	return p.sendWithRetry(ctx, event, currentRetry+1)
}

func (p *Producer[T]) sendWithRetry(ctx context.Context, event T, retryCount int) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Value: body,
		Headers: []kafka.Header{
			{Key: "x-retry-count", Value: []byte(strconv.Itoa(retryCount))},
		},
	})
}

func (p *Producer[T]) Close() error {
	return p.writer.Close()
}
