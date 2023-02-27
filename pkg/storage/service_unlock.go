package storage

import (
	"context"
	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/viper"
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

	opts = badger.DefaultOptions(viper.GetString(storeFilePathFlag))
	opts.Logger = nil
	opts.IndexCacheSize = 128 << 20
	opts.EncryptionKey = key.Bytes()

	if db, err = badger.Open(opts); err != nil {

		log.I.Ln("unlock attempt failed:", err)

		return &UnlockResponse{
			Success: false,
		}, err
	}

	s.success <- true
	isUnlockedChan <- true

	log.I.Ln("unlock successful")

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
