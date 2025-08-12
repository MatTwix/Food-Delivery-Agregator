package clients

import (
	"log/slog"
	"os"

	pb "github.com/MatTwix/Food-Delivery-Agregator/common/proto"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewResraurantServiceClient() pb.RestaurantServiceClient {
	conn, err := grpc.NewClient("restaurants-service:"+config.Cfg.GRPC.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to connect to gRPC server", "error", err)
		os.Exit(1)
	}

	slog.Info("successfully connected to restaurants-service gRPC server")
	return pb.NewRestaurantServiceClient(conn)
}
