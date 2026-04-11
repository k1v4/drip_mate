package kafka

import (
	"context"
	"errors"
	"strconv"

	"github.com/k1v4/drip_mate/internal/modules/notification_service/entity"
	"github.com/k1v4/drip_mate/internal/modules/notification_service/usecase"
	"github.com/k1v4/drip_mate/pkg/logger"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

const RetryHeader = "x-retry-count"

type Consumer struct {
	reader  *kafka.Reader
	handler usecase.IHandler
	l       logger.Logger
}

func NewConsumer(
	reader *kafka.Reader,
	handler usecase.IHandler,
	l logger.Logger,
) *Consumer {
	return &Consumer{
		reader:  reader,
		handler: handler,
		l:       l,
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

		mappedMsg := mapMessage(new(msg))

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

		//nolint
		err = c.handler.Handle(ctx, new(mappedMsg))
		if err != nil {
			// логируем и НЕ коммитим offset
			// сообщение будет перечитано
			c.l.Error(
				ctx,
				"error handling message",
				zap.Error(err),
				zap.String("topic", mappedMsg.Topic),
				zap.Int("partition", mappedMsg.Partition),
				zap.Int64("offset", mappedMsg.Offset),
			)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}

func mapMessage(message *kafka.Message) entity.Message {
	return entity.Message{}
}

func getRetryCount(msg *entity.Message) int {
	if v, ok := msg.Headers[RetryHeader]; ok {
		n, _ := strconv.Atoi(string(v))
		return n
	}
	return 0
}
