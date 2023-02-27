package seed

import (
	"context"
	"git-indra.lan/indra-labs/indra/pkg/rpc"
	"git-indra.lan/indra-labs/indra/pkg/storage"
	"google.golang.org/grpc"
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
		case <-storage.WhenIsLocked():

			log.I.Ln("storage is locked")

			// Run an unlock RPC server
			go rpc.RunWith(
				func(srv *grpc.Server) {
					storage.RegisterUnlockServiceServer(srv, storage.NewUnlockService())
				},
				rpc.WithDisableTunnel(),
			)

		case err := <-rpc.WhenStartFailed():
			startupErrors <- err
		case <-rpc.IsReady():
			log.I.Ln("waiting for unlock")
		case <-storage.WhenIsUnlocked():

			log.I.Ln("restarting rpc server")

			// Shut down unlock RPC server to we can launch the main one
			rpc.Shutdown()

			//go rpc.RunWith(func(srv *grpc.Server) {
			//	chat.RegisterChatServiceServer(srv, &chat.Server{})
			//})

		case <-storage.WhenIsReady():
			break signals
		case <-ctx.Done():
			Shutdown()
			return
		}
	}

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

	rpc.Shutdown()

	err := storage.Shutdown()
	check(err)

	isShutdownChan <- true
}
