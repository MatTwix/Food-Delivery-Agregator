package api

import (
	"context"
	"log/slog"

	pb "github.com/MatTwix/Food-Delivery-Agregator/common/proto"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/store"
)

type GrpcServer struct {
	pb.UnimplementedRestaurantServiceServer
	menuItemStore *store.MenuItemStore
}

func NewGrpcServer(s *store.MenuItemStore) *GrpcServer {
	return &GrpcServer{menuItemStore: s}
}

func (s *GrpcServer) GetMenuItems(ctx context.Context, req *pb.GetMenuItemsRequest) (*pb.GetMenuItemsResponse, error) {
	slog.Info("received gRPC request", "menu_items", req.MenuItemIds)

	items, err := s.menuItemStore.GetByIDs(ctx, req.MenuItemIds)
	if err != nil {
		slog.Error("failed to get menu items from store", "error", err)
		return nil, err
	}

	return &pb.GetMenuItemsResponse{MenuItems: items}, nil
}
