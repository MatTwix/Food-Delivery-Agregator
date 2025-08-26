package scheduler

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	pb "github.com/MatTwix/Food-Delivery-Agregator/common/proto"
	"github.com/MatTwix/Food-Delivery-Agregator/scheduler-service/messaging"
)

type RequestCourierJob struct {
	ordersClient pb.OrderServiceClient
	producer     *messaging.Producer
}

func NewRequestCourierJob(ordersClient pb.OrderServiceClient, p *messaging.Producer) *RequestCourierJob {
	return &RequestCourierJob{
		ordersClient: ordersClient,
		producer:     p,
	}
}

func (j *RequestCourierJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	slog.Info("requesting orders for retry")

	resp, err := j.ordersClient.GetRetryOrders(ctx, &pb.GetRetryOrdersRequest{Status: "no_couriers_available"})
	if err != nil {
		slog.Error("failed to fetch orders", "error", err)
		return
	}

	if len(resp.Orders) == 0 {
		slog.Info("there are no orders to search couriers for")
	}

	for _, order := range resp.Orders {
		slog.Info("processing retry", "orderID", order.Id)

		event := struct {
			OrderID string `json:"order_id"`
		}{OrderID: order.Id}

		eventBody, err := json.Marshal(event)
		if err != nil {
			slog.Error("failed to marshal event", "orderID", order.Id, "error", err)
			continue
		}

		err = j.producer.Produce(ctx, messaging.CourierRequestedTopic, []byte(order.Id), eventBody)
		if err != nil {
			slog.Error("failed to send courier requested event", "orderID", order.Id, "error", err)
		}
	}
}
