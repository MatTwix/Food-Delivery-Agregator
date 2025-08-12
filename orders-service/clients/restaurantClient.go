package clients

import (
	"log"

	pb "github.com/MatTwix/Food-Delivery-Agregator/common/proto"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewResraurantServiceClient() pb.RestaurantServiceClient {
	conn, err := grpc.NewClient("restaurants-service:"+config.Cfg.GRPC.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Error connecting to gRPC server: %v", err)
	}

	log.Println("Successfully connected to restaurants-service gRPC server")
	return pb.NewRestaurantServiceClient(conn)
}
