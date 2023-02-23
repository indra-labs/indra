package client

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"strings"
)

var (
	rpcEndpointIp   string = "192.168.37.1"
	rpcEndpointPort string = "80"
)

func DialContext(ctx context.Context, target string, opts ...DialOption) (conn *grpc.ClientConn, err error) {

	if strings.HasPrefix(target, "unix://") {
		return grpc.Dial(
			target,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
	}

	if !strings.HasPrefix(target, "noise://") {
		return nil, errors.New("Unsupported protocol. Only unix:// or noise://")
	}

	dialOpts := &dialOptions{peerRPCIP: "192.168.37.2"}

	for _, opt := range opts {
		opt.apply(dialOpts)
	}

	getNetworkInstance(dialOpts)

	return grpc.DialContext(ctx,
		rpcEndpointIp+":"+rpcEndpointPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, address string) (net.Conn, error) {
			return network.DialContext(ctx, "tcp4", address)
		}))
}

func Dial(target string, opts ...DialOption) (conn *grpc.ClientConn, err error) {
	return DialContext(context.Background(), target, opts...)
}
