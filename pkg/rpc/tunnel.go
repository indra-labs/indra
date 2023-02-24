package rpc

import (
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/tun/netstack"
	"google.golang.org/grpc"
	"net"
	"net/netip"
)

const NullPort = 0

var (
	isTunnelEnabled bool = false
)

var (
	network *netstack.Net
	tunnel  tun.Device
	tcpSock net.Listener
)

var (
	tunKey       *RPCPrivateKey
	tunWhitelist []RPCPublicKey
	tunnelPort   int = 0
	tunnelMTU    int = 1420
)

func enableTunnel() {

	var err error

	isTunnelEnabled = true

	if tunnel, network, err = netstack.CreateNetTUN([]netip.Addr{deviceRPCIP}, []netip.Addr{}, tunnelMTU); check(err) {
		startupErrors <- err
		return
	}

	dev = device.NewDevice(tunnel, conn.NewDefaultBind(), device.NewLogger(device.LogLevelError, "server "))
}

func startTunnel(srv *grpc.Server) (err error) {

	if !isTunnelEnabled {
		return
	}

	configureDevice()

	if err = dev.Up(); check(err) {
		startupErrors <- err
		return
	}

	if tcpSock, err = network.ListenTCPAddrPort(netip.AddrPortFrom(deviceRPCIP, deviceRPCPort)); check(err) {
		startupErrors <- err
		return
	}

	go srv.Serve(tcpSock)

	return
}

func stopTunnel() (err error) {

	if !isTunnelEnabled {
		return
	}

	if err = tcpSock.Close(); check(err) {
		// continue
	}

	dev.Close()

	return
}
