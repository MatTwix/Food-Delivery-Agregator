package api

import (
	"context"
	"log"

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
	log.Printf("Received gRPC request for menu items: %v", req.MenuItemIds)

	items, err := s.menuItemStore.GetByIDs(ctx, req.MenuItemIds)
	if err != nil {
		log.Printf("Error getting menu items from store: %v", err)
		return nil, err
	}

	return &pb.GetMenuItemsResponse{MenuItems: items}, nil
}
