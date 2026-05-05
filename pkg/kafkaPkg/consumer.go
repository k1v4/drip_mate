package kafkaPkg

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	totalEntity "github.com/k1v4/drip_mate/internal/modules/notification_service/entity"
	"github.com/k1v4/drip_mate/pkg/logger"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

const RetryHeader = "x-retry-count"

type Consumer[T any] struct {
	reader   KafkaReader
	producer RetryProducer[T]
	handler  func(context.Context, *T) error
	l        logger.Logger
}

func NewConsumer[T any](
	reader KafkaReader,
	producer RetryProducer[T],
	handler func(context.Context, *T) error,
	l logger.Logger,
) *Consumer[T] {
	return &Consumer[T]{
		reader:   reader,
		producer: producer,
		handler:  handler,
		l:        l,
	}
}

func (c *Consumer[T]) Run(ctx context.Context) error {
	c.l.Info(ctx, "consumer started, waiting for messages",
		zap.String("topic", c.reader.Config().Topic),
		zap.String("groupID", c.reader.Config().GroupID),
	)

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				c.l.Info(ctx, "consumer stopped")
				return nil
			}
			c.l.Error(ctx, "read message error", zap.Error(err))
			return err
		}

		c.l.Info(ctx, "received message",
			zap.String("topic", msg.Topic),
			zap.String("value", string(msg.Value)),
		)

		mappedMsg := mapMessage(new(msg))

		retryCount := getRetryCount(new(mappedMsg))

		if retryCount > 5 {
			c.l.Error(ctx, "retry limit exceeded, skipping message",
				zap.String("topic", mappedMsg.Topic),
				zap.Int("partition", mappedMsg.Partition),
				zap.Int64("offset", mappedMsg.Offset),
			)
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				return err
			}
			continue
		}

		var event T
		if err := json.Unmarshal(mappedMsg.Value, &event); err != nil {
			c.l.Error(ctx, "unmarshal error", zap.Error(err))
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				return err
			}
			continue
		}

		// Вызываем хендлер
		if err := c.handler(ctx, &event); err != nil {
			c.l.Error(ctx, "error handling message", zap.Error(err))

			// Отправляем в ретрай тот же тип события
			if retryErr := c.producer.Retry(ctx, event, retryCount); retryErr != nil {
				c.l.Error(ctx, "failed to retry", zap.Error(retryErr))
			}

			_ = c.reader.CommitMessages(ctx, msg)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}

func mapMessage(message *kafka.Message) totalEntity.Message {
	headers := make(map[string][]byte, len(message.Headers))
	for _, h := range message.Headers {
		headers[h.Key] = h.Value
	}

	return totalEntity.Message{
		Key:       message.Key,
		Value:     message.Value,
		Headers:   headers,
		Topic:     message.Topic,
		Partition: message.Partition,
		Offset:    message.Offset,
	}
}

func getRetryCount(msg *totalEntity.Message) int {
	if v, ok := msg.Headers[RetryHeader]; ok {
		n, _ := strconv.Atoi(string(v))
		return n
	}
	return 0
}
