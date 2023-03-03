package rpc

import (
	"google.golang.org/grpc"
	"sync"
)

var (
	server *grpc.Server
	o      *serverOptions
)

var (
	inUse     sync.Mutex
	isRunning bool
)

func RunWith(r func(srv *grpc.Server), opts ...ServerOption) {

	if !inUse.TryLock() {
		log.E.Ln("rpc server is in use.")
		return
	}

	log.I.Ln("initializing the rpc server")

	o = &serverOptions{false, &storeMem{}}

	for _, opt := range opts {
		opt.apply(o)
	}

	server = grpc.NewServer()

	configureUnixSocket()
	configureTunnel()

	r(server)

	go start()
}

func start() {

	log.I.Ln("starting rpc server")

	var err error

	createTunnel()

	if err = startTunnel(server); check(err) {
		startupErrors <- err
		return
	}

	if err = startUnixSocket(server); check(err) {
		startupErrors <- err
		return
	}

	isRunning = true

	log.I.Ln("rpc server is ready")
	isReady <- true
}

func Shutdown() {

	if !isRunning {
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

	isRunning = false

	inUse.Unlock()

	log.I.Ln("- rpc server shutdown completed")
}
