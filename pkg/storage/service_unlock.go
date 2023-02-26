package storage

import (
	"context"
	"github.com/dgraph-io/badger/v3"
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

	opts.EncryptionKey = key.Bytes()

	if db, err = badger.Open(opts); check(err) {
		return &UnlockResponse{
			Success: false,
		}, err
	}

	s.success <- true

	return &UnlockResponse{
		Success: true,
	}, nil
}

func (s *Service) mustEmbedUnimplementedUnlockServiceServer() {}

func NewUnlockService() *Service {
	return &Service{
		success: make(chan bool, 1),
	}
}
