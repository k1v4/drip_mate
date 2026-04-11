package kafkaPkg

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/k1v4/drip_mate/internal/entity"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  topic,
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
		},
	}
}

func (p *Producer) SendNotification(ctx context.Context, event entity.NotificationEvent) error {
	return p.sendWithRetry(ctx, event, 0)
}

func (p *Producer) Retry(ctx context.Context, event entity.NotificationEvent, currentRetry int) error {
	return p.sendWithRetry(ctx, event, currentRetry+1)
}

func (p *Producer) sendWithRetry(ctx context.Context, event entity.NotificationEvent, retryCount int) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal notification event: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Value: body,
		Headers: []kafka.Header{
			{
				Key:   RetryHeader,
				Value: []byte(strconv.Itoa(retryCount)),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("write kafka message: %w", err)
	}

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
