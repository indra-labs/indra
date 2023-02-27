package seed

import (
	"context"
	"git-indra.lan/indra-labs/indra/pkg/storage"
	"sync"
)

var (
	inUse sync.Mutex
)

var (
	startupErrors  = make(chan error, 32)
	isReadyChan    = make(chan bool, 1)
	isShutdownChan = make(chan bool, 1)
)

func CantStart() chan error {
	return startupErrors
}

func IsReady() chan bool {
	return isReadyChan
}

func IsShutdown() chan bool {
	return isShutdownChan
}

func Shutdown() {

	log.I.Ln("shutting down seed")

	err := storage.Shutdown()
	check(err)

	isShutdownChan <- true
}

func Run(ctx context.Context) {

	if !inUse.TryLock() {
		log.E.Ln("seed is in use")
		return
	}

	log.I.Ln("running seed")

	var err error

	go storage.Run(ctx)

signals:
	for {
		select {
		case err = <-CantStart():
			log.E.Ln("startup error:", err)
			return
		case <-storage.IsLocked():

			log.I.Ln("storage is locked, waiting for unlock")

			//go rpc.RunWith(ctx, func(srv *grpc.Server) {
			//	storage.RegisterUnlockServiceServer(srv, storage.NewUnlockService())
			//})

			// Run RPC unlock
		case <-storage.IsReady():
			break signals
		case <-ctx.Done():
			Shutdown()
			return
		}
	}

	log.I.Ln("seed is ready")

	isReadyChan <- true
}
