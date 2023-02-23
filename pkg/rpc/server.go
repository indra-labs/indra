package rpc

import (
	"context"
	"google.golang.org/grpc"
)

func init() {
	server = grpc.NewServer()
}

var (
	server *grpc.Server
)

var (
	isReady       = make(chan bool)
	startupErrors = make(chan error)
)

func IsReady() chan bool {
	return isReady
}

func CantStart() chan error {
	return startupErrors
}

func Register(r func(srv *grpc.Server)) {
	r(server)
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

	defer ctx.Done()

	log.I.Ln("shutting down rpc server")

	stopUnixSocket()

	server.Stop()
}
