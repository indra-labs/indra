package storage

import (
	"context"
	"github.com/dgraph-io/badger/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	success chan bool
}

func (s *Service) IsSuccessful() chan bool {
	return s.success
}

func (s *Service) Unlock(ctx context.Context, req *UnlockRequest) (res *UnlockResponse, err error) {

	var key Key

	key.Decode(req.Key)

	if db, err = badger.Open(opts); check(err) {
		return &UnlockResponse{
			Success: false,
		}, err
	}

	return nil, status.Errorf(codes.Unimplemented, "method Unlock not implemented")
}

func (s *Service) mustEmbedUnimplementedUnlockServiceServer() {}

func NewUnlockService() UnlockServiceServer {
	return &Service{
		success: make(chan bool, 1),
	}
}
