package rpc

import (
	"context"
	"errors"
	"golang.zx2c4.com/wireguard/tun/netstack"
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

	dialOpts := &dialOptions{
		endpoint:    EndpointString(target),
		rpcEndpoint: EndpointString("192.168.37.1:80"),
		peerRPCIP:   "192.168.37.2",
		mtu:         1420,
	}

	for _, opt := range opts {
		opt.apply(dialOpts)
	}

	var network *netstack.Net

	if network, err = getNetworkInstance(dialOpts); check(err) {
		return
	}

	return grpc.DialContext(ctx,
		dialOpts.rpcEndpoint.String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, address string) (net.Conn, error) {
			return network.DialContext(ctx, "tcp4", address)
		}))
}

func Dial(target string, opts ...DialOption) (conn *grpc.ClientConn, err error) {
	return DialContext(context.Background(), target, opts...)
}
