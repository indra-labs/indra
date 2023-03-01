package rpc

import (
	"google.golang.org/grpc"
)

var (
	server *grpc.Server
)

var (
	running bool
)

func RunWith(r func(srv *grpc.Server), opts ...ServerOption) {

	log.I.Ln("initializing the rpc server")

	serverOpts := serverOptions{}

	for _, opt := range opts {
		opt.apply(&serverOpts)
	}

	server = grpc.NewServer()

	configureUnixSocket()
	configureTunnel()

	r(server)

	log.I.Ln("starting rpc server")

	go Start()
}

func Start() {

	var err error

	if err = startTunnel(server); check(err) {
		startupErrors <- err
	}

	if err = startUnixSocket(server); check(err) {
		startupErrors <- err
	}

	running = true

	log.I.Ln("rpc server is ready")

	isReady <- true
}

func Shutdown() {

	if !running {
		return
	}

	log.I.Ln("shutting down rpc server")

	server.GracefulStop()

	var err error

	//err = stopUnixSocket()
	//
	//check(err)

	err = stopTunnel()

	check(err)

	running = false

	log.I.Ln("- rpc server shutdown completed")
}
