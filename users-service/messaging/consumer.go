package messaging

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/MatTwix/Food-Delivery-Agregator/users-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/store"
	"github.com/segmentio/kafka-go"
)

type RefreshTokenDeletionEvent struct {
	TokenID string `json:"token_id"`
}

func StartConsumers(ctx context.Context, tokenStore *store.TokenStore) {
	go startTopicConsumer(ctx, RefreshTokenDeletionRequestedTopic, config.Cfg.Kafka.GroupIDs.Tokens, func(ctx context.Context, msg kafka.Message) {
		handleTokenDeletionRequested(ctx, msg, tokenStore)
	})
}

func startTopicConsumer(ctx context.Context, topic, groupID string, handler func(ctx context.Context, msg kafka.Message)) {
	if config.Cfg.Kafka.Brokers == "" {
		slog.Error("KAFKA_BROKERS environment variable is not set")
		os.Exit(1)
	}

	brokers := strings.Split(config.Cfg.Kafka.Brokers, ",")

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		CommitInterval: 1 * time.Second,
		StartOffset:    kafka.LastOffset,
	})

	slog.Info("Starting Kafka consumer", "topic", topic, "group_id", groupID)

	defer r.Close()

	for {
		select {
		case <-ctx.Done():
			slog.Info("stopping consumer due to context cancellation", "topic", topic)
			return
		default:
			m, err := r.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					slog.Info("context cancelled, stopping consumer", "topic", topic)
					return
				}
				slog.Error("failed to read message", "topic", topic, "error", err)
				continue
			}
			slog.Info("processing message", "topic", topic)
			handler(ctx, m)

			if err := r.CommitMessages(ctx, m); err != nil {
				slog.Error("failed to commit message offset", "error", err)
			}
		}
	}
}

func handleTokenDeletionRequested(ctx context.Context, msg kafka.Message, tokenStore *store.TokenStore) {
	slog.Info("handling event", "event", RefreshTokenDeletionRequestedTopic)

	tokenID := string(msg.Key)

	if err := tokenStore.DeleteRefreshTokenByID(ctx, tokenID); err != nil {
		slog.Error("failed to delete token", "tokenID", tokenID, "error", err)
		return
	}

	slog.Info("successfully deleted token from db", "tokenID", tokenID)
}
