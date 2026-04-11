package kafkaPkg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/k1v4/drip_mate/internal/entity"
	notificationEntity "github.com/k1v4/drip_mate/internal/modules/notification_service/entity"
	"github.com/k1v4/drip_mate/internal/modules/notification_service/usecase"
	"github.com/k1v4/drip_mate/pkg/logger"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

const RetryHeader = "x-retry-count"

type Consumer struct {
	reader   *kafka.Reader
	producer *Producer
	handler  usecase.IHandler
	l        logger.Logger
}

func NewConsumer(
	reader *kafka.Reader,
	handler usecase.IHandler,
	producer *Producer,
	l logger.Logger,
) *Consumer {
	return &Consumer{
		reader:   reader,
		handler:  handler,
		producer: producer,
		l:        l,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}

		mappedMsg := mapMessage(&msg)
		c.l.Info(ctx, fmt.Sprintf("Message received: %v", mappedMsg))

		if getRetryCount(new(mappedMsg)) > 5 {
			c.l.Error(
				ctx,
				"retry limit exceeded, skipping message",
				zap.String("topic", mappedMsg.Topic),
				zap.Int("partition", mappedMsg.Partition),
				zap.Int64("offset", mappedMsg.Offset),
			)

			// принудительно коммитим offset
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				return err
			}

			continue
		}

		var event entity.NotificationEvent
		if err := json.Unmarshal(mappedMsg.Value, &event); err != nil {
			return fmt.Errorf("unmarshal event: %w", err)
		}

		err = c.handler.Handle(ctx, &event)
		if err != nil {
			c.l.Error(ctx, "error handling message", zap.Error(err))

			// 4. при ошибке — переотправляем с incremented retry count
			retryCount := getRetryCount(&mappedMsg)
			if retryErr := c.producer.Retry(ctx, event, retryCount); retryErr != nil {
				c.l.Error(ctx, "failed to retry", zap.Error(retryErr))
			}

			// коммитим оригинал чтобы не читать его снова
			_ = c.reader.CommitMessages(ctx, msg)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}

func mapMessage(message *kafka.Message) notificationEntity.Message {
	headers := make(map[string][]byte, len(message.Headers))
	for _, h := range message.Headers {
		headers[h.Key] = h.Value
	}

	return notificationEntity.Message{
		Key:       message.Key,
		Value:     message.Value,
		Headers:   headers,
		Topic:     message.Topic,
		Partition: message.Partition,
		Offset:    message.Offset,
	}
}

func getRetryCount(msg *notificationEntity.Message) int {
	if v, ok := msg.Headers[RetryHeader]; ok {
		n, _ := strconv.Atoi(string(v))
		return n
	}
	return 0
}
