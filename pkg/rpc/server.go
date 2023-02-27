package rpc

import (
	"context"
	"google.golang.org/grpc"
)

func init() {
	server = grpc.NewServer()
}

var (
	server        *grpc.Server
	startupErrors = make(chan error, 128)
	isReady       = make(chan bool, 1)
)

func RunWith(ctx context.Context, r func(srv *grpc.Server)) {

	log.I.Ln("initializing the rpc server")

	configureUnixSocket()
	configureTunnel()

	r(server)

	log.I.Ln("starting rpc server")

	go Start(ctx)
}

func CantStart() chan error {
	return startupErrors
}

func IsReady() chan bool {
	return isReady
}

func Start(ctx context.Context) {

	var err error

	if err = startUnixSocket(server); check(err) {
		startupErrors <- err
	}

	if err = startTunnel(server); check(err) {
		startupErrors <- err
	}

	isReady <- true

	select {
	case <-ctx.Done():
		Shutdown(context.Background())
	}
}

func Shutdown(ctx context.Context) {

	log.I.Ln("shutting down rpc server")

	stopUnixSocket()
	stopTunnel()

	server.Stop()
}
