// Package seed provides an implementation of an indra seed server, its purpose to be a rendezvous point for non-routeable clients as well as gathering and distributing current peer information metadata.
package seed

import (
	"context"
	"github.com/indra-labs/indra/pkg/p2p"
	"github.com/indra-labs/indra/pkg/rpc"
	"github.com/indra-labs/indra/pkg/storage"
	"github.com/spf13/viper"
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

	opts := []rpc.ServerOption{
		rpc.WithUnixPath(
			viper.GetString(rpc.UnixPathFlag),
		),
		rpc.WithStore(
			&rpc.BadgerStore{storage.DB()},
		),
	}

	if viper.GetBool(rpc.TunEnableFlag) {
		opts = append(opts,
			rpc.WithTunOptions(
				viper.GetUint16(rpc.TunPortFlag),
				viper.GetStringSlice(rpc.TunPeersFlag),
			))
	}

	services := func(srv *grpc.Server) {
		chat.RegisterChatServiceServer(srv, &chat.Server{})
	}

	go rpc.RunWith(services, opts...)

signals:
	for {
		select {
		case <-rpc.IsConfigured():
			// We need to get the randomly generated port
			viper.Set(rpc.TunPortFlag, rpc.Options().GetTunPort())
		case err := <-rpc.WhenStartFailed():
			log.E.Ln("rpc can't start:", err)
			startupErrors <- err
			return
		case <-rpc.IsReady():
			break signals
		case <-ctx.Done():
			Shutdown()
			return
		}
	}

	//
	// Ready!
	//

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
