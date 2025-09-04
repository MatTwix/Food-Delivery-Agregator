package scheduler

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	pb "github.com/MatTwix/Food-Delivery-Agregator/common/proto"
	"github.com/MatTwix/Food-Delivery-Agregator/scheduler-service/messaging"
)

type DeleteExpiredTokensJob struct {
	spec        string
	usersClient pb.UserServiceClient
	producer    *messaging.Producer
}

func NewDeleteExpiredTokensJob(usersClient pb.UserServiceClient, p *messaging.Producer) *DeleteExpiredTokensJob {
	return &DeleteExpiredTokensJob{
		spec:        "@every 15m",
		usersClient: usersClient,
		producer:    p,
	}
}

func (j *DeleteExpiredTokensJob) Spec() string {
	return j.spec
}

func (j *DeleteExpiredTokensJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	slog.Info("requesting expired tokens to delete")

	resp, err := j.usersClient.GetExpiredRefreshTokensIDs(ctx, &pb.GetExpiredRefreshTokensIDsRequest{
		BeforeDate: time.Now().Unix(),
	})
	if err != nil {
		slog.Error("failed to getch tokens", "error", err)
		return
	}

	if len(resp.TokenIds) == 0 {
		slog.Info("there are no expired tokens to delete")
		return
	}

	deletedTokensAmount := 0

	for _, tokenID := range resp.TokenIds {
		event := struct {
			TokenID string `json:"token_id"`
		}{TokenID: tokenID}

		eventBody, err := json.Marshal(event)
		if err != nil {
			slog.Error("failed to marshal event", "token_id", tokenID, "error", err)
			continue
		}

		err = j.producer.Produce(ctx, messaging.RefreshTokenDeletionRequestedTopic, []byte(tokenID), eventBody)
		if err != nil {
			slog.Error("failed to send token deletion requested event", "tokenID", tokenID, "error", err)
		}

		deletedTokensAmount++
	}

	slog.Info("expired tokens deleted", "amount", deletedTokensAmount)
}
