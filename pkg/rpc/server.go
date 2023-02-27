package rpc

import (
	"google.golang.org/grpc"
)

func init() {
	server = grpc.NewServer()
}

var (
	server *grpc.Server
)

func RunWith(r func(srv *grpc.Server), opts ...ServerOption) {

	log.I.Ln("initializing the rpc server")

	serverOpts := serverOptions{}

	for _, opt := range opts {
		opt.apply(&serverOpts)
	}
	
	configureUnixSocket()
	configureTunnel()

	r(server)

	log.I.Ln("starting rpc server")

	go Start()
}

func Start() {

	var err error

	if err = startUnixSocket(server); check(err) {
		startupErrors <- err
	}

	if err = startTunnel(server); check(err) {
		startupErrors <- err
	}

	log.I.Ln("rpc server is ready")

	isReady <- true
}

func Shutdown() {

	log.I.Ln("shutting down rpc server")

	stopUnixSocket()
	stopTunnel()

	server.Stop()
}
