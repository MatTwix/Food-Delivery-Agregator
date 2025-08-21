package api

import (
	"context"

	pb "github.com/MatTwix/Food-Delivery-Agregator/common/proto"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/store"
)

type OrderGRPCServer struct {
	pb.UnimplementedOrderServiceServer
	orderStore *store.OrderStore
}

func NewOrderGRPCServer(orderStore *store.OrderStore) *OrderGRPCServer {
	return &OrderGRPCServer{
		orderStore: orderStore,
	}
}

func (s *OrderGRPCServer) GetOrderOwner(ctx context.Context, req *pb.GetOrderOwnerRequest) (*pb.GetOrderOwnerResponce, error) {
	userID, err := s.orderStore.GetOwnerID(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}

	return &pb.GetOrderOwnerResponce{
		UserId: userID,
	}, nil
}
