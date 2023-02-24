package rpc

import (
	"google.golang.org/grpc"
	"net"
	"os"
)

var (
	isUnixSockEnabled bool = false
	unixSock          net.Listener
	unixPath          = "/tmp/indra.sock"
)

func startUnixSocket(srv *grpc.Server) (err error) {

	if !isUnixSockEnabled {
		return
	}

	if unixSock, err = net.Listen("unix", unixPath); err != nil {
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
		if err = unixSock.Close(); check(err) {
			// continue
		}
	}

	os.Remove(unixPath)

	return
}
