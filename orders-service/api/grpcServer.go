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

func (s *OrderGRPCServer) GetOrderStatus(ctx context.Context, req *pb.GetOrderStatusRequest) (*pb.GetOrderStatusResponce, error) {
	status, err := s.orderStore.GetStatus(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}

	return &pb.GetOrderStatusResponce{
		Status: status,
	}, nil
}

func (s *OrderGRPCServer) GetRetryOrders(ctx context.Context, req *pb.GetRetryOrdersRequest) (*pb.GetRetryOrdersResponce, error) {
	orders, err := s.orderStore.GetForRetry(ctx, req.Status, req.NextRetryAtLte, req.Limit)
	if err != nil {
		return nil, err
	}

	var pbOrders []*pb.OrderLite
	for _, order := range orders {
		pbOrders = append(pbOrders, &pb.OrderLite{
			Id:            order.ID,
			RetryCount:    int32(order.RetryCount),
			MaxRetryCount: int32(order.MaxRetryCount),
			NextRetryAt:   order.NextRetryAt.Unix(),
		})
	}

	return &pb.GetRetryOrdersResponce{Orders: pbOrders}, nil
}
