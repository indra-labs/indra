package storage

import (
	"context"
)

type Service struct{}

func (s *Service) Unlock(ctx context.Context, req *UnlockRequest) (res *UnlockResponse, err error) {

	key.Decode(req.Key)

	isUnlocked, err := attempt_unlock()

	if !isUnlocked {

		log.I.Ln("unlock attempt failed:", err)

		return &UnlockResponse{Success: false}, err
	}

	log.I.Ln("successfully unlocked database")
	isUnlockedChan <- true

	return &UnlockResponse{Success: true}, nil
}

func (s *Service) mustEmbedUnimplementedUnlockServiceServer() {}

func NewUnlockService() *Service { return &Service{} }
