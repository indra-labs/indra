package seed

import (
	"context"
	"git-indra.lan/indra-labs/indra/pkg/storage"
	"sync"
)

var (
	inUse sync.Mutex
)

func Run(ctx context.Context) {

	if !inUse.TryLock() {
		log.E.Ln("seed is in use")
		return
	}

	log.I.Ln("running seed")

	go storage.Run()

signals:
	for {
		select {
		case err := <-storage.WhenStartupFailed():
			log.E.Ln("storage can't start:", err)
			startupErrors <- err
			return
		case <-storage.WhenIsReady():

			//log.I.Ln("shutting down rpc server, with unlock service")
			//
			//rpc.Shutdown()

			break signals
		//case <-storage.WhenIsLocked():
		//
		//	log.I.Ln("running rpc server, with unlock service")
		//
		//	go rpc.RunWith(
		//		func(srv *grpc.Server) {
		//			storage.RegisterUnlockServiceServer(srv, storage.NewUnlockService())
		//		},
		//		rpc.WithDisableTunnel(),
		//	)
		//
		//case err := <-rpc.WhenStartFailed():
		//	startupErrors <- err
		//case <-rpc.IsReady():
		//	log.I.Ln("... awaiting unlock over rpc")
		case <-ctx.Done():
			Shutdown()
			return
		}
	}

	// Startup all RPC services

	// Startup P2P services

	log.I.Ln("seed is ready")

	isReadyChan <- true

	select {
	case <-ctx.Done():
		Shutdown()
		return
	}
}

func Shutdown() {

	log.I.Ln("shutting down seed")

	err := storage.Shutdown()
	check(err)

	log.I.Ln("seed shutdown completed")

	isShutdownChan <- true
}
