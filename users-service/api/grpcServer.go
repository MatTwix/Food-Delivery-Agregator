package api

import (
	"context"
	"time"

	pb "github.com/MatTwix/Food-Delivery-Agregator/common/proto"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/store"
)

type UserGRPCServer struct {
	pb.UnimplementedUserServiceServer
	tokenStore *store.TokenStore
}

func NewUserGRPCServer(tokenStore *store.TokenStore) *UserGRPCServer {
	return &UserGRPCServer{
		tokenStore: tokenStore,
	}
}

func (s *UserGRPCServer) GetExpiredRefreshTokensIDs(ctx context.Context, req *pb.GetExpiredRefreshTokensIDsRequest) (*pb.GetExpiredRefreshTokensIDsResponce, error) {
	tokenIDs, err := s.tokenStore.GetExpiredTokensIDs(ctx, time.Unix(req.BeforeDate, 0))
	if err != nil {
		return nil, err
	}

	return &pb.GetExpiredRefreshTokensIDsResponce{TokenIds: tokenIDs}, nil
}
