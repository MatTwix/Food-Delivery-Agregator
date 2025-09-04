package clients

import (
	"log/slog"
	"os"

	pb "github.com/MatTwix/Food-Delivery-Agregator/common/proto"
	"github.com/MatTwix/Food-Delivery-Agregator/scheduler-service/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewUsersServiceClient() pb.UserServiceClient {
	conn, err := grpc.NewClient("users-service:"+config.Cfg.GRPC.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to connect to gRPC server", "error", err)
		os.Exit(1)
	}

	slog.Info("successfully connected to users-service gRPC server")
	return pb.NewUserServiceClient(conn)
}
