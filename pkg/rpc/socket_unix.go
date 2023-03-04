package rpc

import (
	"google.golang.org/grpc"
	"net"
	"os"
)

const unixPathDefault = "/tmp/indra.sock"

var (
	isUnixSockEnabled bool = false
	unixSock          net.Listener
)

func startUnixSocket(srv *grpc.Server) (err error) {

	if !isUnixSockEnabled {
		return
	}

	if unixSock, err = net.Listen("unix", o.unixPath); err != nil {
		return
	}

	go srv.Serve(unixSock)

	return
}

func stopUnixSocket() (err error) {

	if !isUnixSockEnabled {
		return
	}

	if unixSock != nil {
		if err = unixSock.Close(); err != nil {
			// continue
		}
	}

	os.Remove(o.unixPath)

	return
}
