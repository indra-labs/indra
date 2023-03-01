package seed

import (
	"context"
	"git-indra.lan/indra-labs/indra/pkg/p2p"
	"git-indra.lan/indra-labs/indra/pkg/rpc"
	"git-indra.lan/indra-labs/indra/pkg/storage"
	"github.com/tutorialedge/go-grpc-tutorial/chat"
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

	//
	// Storage
	//

	go storage.Run()

	select {
	case err := <-storage.WhenStartFailed():
		log.E.Ln("storage can't start:", err)
		startupErrors <- err
		return
	case <-storage.WhenReady():
		// continue
	case <-ctx.Done():
		Shutdown()
		return
	}

	//
	// P2P
	//

	go p2p.Run()

	select {
	case err := <-p2p.WhenStartFailed():
		log.E.Ln("p2p can't start:", err)
		startupErrors <- err
		return
	case <-p2p.WhenReady():
		// continue
	case <-ctx.Done():
		Shutdown()
		return
	}

	//
	// RPC
	//

	//
	// Ready!
	//

	go rpc.RunWith(func(srv *grpc.Server) {
		chat.RegisterChatServiceServer(srv, &chat.Server{})
	})

	select {
	case err := <-rpc.WhenStartFailed():
		log.E.Ln("rpc can't start:", err)
		startupErrors <- err
		return
	case <-rpc.IsReady():
		// continue
	case <-ctx.Done():
		Shutdown()
		return
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

	var err error

	err = p2p.Shutdown()
	check(err)

	err = storage.Shutdown()
	check(err)

	log.I.Ln("seed shutdown completed")

	isShutdownChan <- true
}
