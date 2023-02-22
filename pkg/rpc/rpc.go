package rpc

import (
	"context"
	"git-indra.lan/indra-labs/indra"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"github.com/multiformats/go-multiaddr"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/tun/netstack"
	"google.golang.org/grpc"
	"net"
	"net/netip"
	"os"
	"strconv"
)

const NullPort = 0

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

var (
	config = rpcConfig{
		Key:            &nullRPCPrivateKey,
		ListenPort:     NullPort,
		Peer_Whitelist: []RPCPublicKey{},
		IP_Whitelist:   []multiaddr.Multiaddr{},
	}
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

var (
	deviceIP   netip.Addr = netip.MustParseAddr("192.168.4.28")
	devicePort int        = 0
	deviceMTU  int        = 1420
)

var (
	dev      *device.Device
	network  *netstack.Net
	tunnel   tun.Device
	unixSock net.Listener
	tcpSock  net.Listener
	server   *grpc.Server
)

func init() {
	server = grpc.NewServer()
}

func Server() *grpc.Server {
	return server
}

func Start(ctx context.Context) {

	var err error
	var config = config

	// Initializing the tunnel
	if tunnel, network, err = netstack.CreateNetTUN([]netip.Addr{deviceIP}, []netip.Addr{}, deviceMTU); check(err) {
		startupErrors <- err
		return
	}

	dev = device.NewDevice(tunnel, conn.NewDefaultBind(), device.NewLogger(device.LogLevelError, "server "))

	dev.SetPrivateKey(config.Key.AsDeviceKey())
	dev.IpcSet("listen_port=" + strconv.Itoa(int(config.ListenPort)))

	for _, peer_whitelist := range config.Peer_Whitelist {

		deviceConf := "" +
			"public_key=" + peer_whitelist.HexString() + "\n" +
			"allowed_ip=" + DefaultClientIPAddr.String() + "/32\n"

		if err = dev.IpcSet(deviceConf); check(err) {
			startupErrors <- err
			return
		}
	}

	if err = dev.Up(); check(err) {
		startupErrors <- err
		return
	}

	if unixSock, err = net.Listen("unix", config.UnixPath); check(err) {
		startupErrors <- err
		return
	}

	go server.Serve(unixSock)

	if tcpSock, err = network.ListenTCPAddrPort(netip.AddrPortFrom(deviceIP, 80)); check(err) {
		startupErrors <- err
		return
	}

	go server.Serve(tcpSock)

	//network.ListenPing(netstack.PingAddrFromAddr(deviceIP))

	isReady <- true

	select {
	case <-ctx.Done():
		Shutdown()
	}
}

func Shutdown() {

	log.I.Ln("shutting down rpc server")

	if unixSock != nil {

		unixSock.Close()

		os.Remove(config.UnixPath)
	}

	if tcpSock != nil {
		tcpSock.Close()
	}

	if dev != nil {
		dev.Close()
	}

	server.Stop()
}
