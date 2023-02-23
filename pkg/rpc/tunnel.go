package rpc

import (
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/tun/netstack"
	"net"
	"net/netip"
	"strconv"
)

const NullPort = 0

var (
	isTunnelEnabled bool = false
)

var (
	deviceRPCIP   netip.Addr = netip.MustParseAddr("192.168.4.28")
	deviceRPCPort uint16     = 80
	devicePort    int        = 0
	deviceMTU     int        = 1420
)

var (
	dev     *device.Device
	network *netstack.Net
	tunnel  tun.Device
	tcpSock net.Listener
)

func enableTunnel() {
	isTunnelEnabled = true
}

func startTunnel() (err error) {

	if !isTunnelEnabled {
		return
	}

	if tunnel, network, err = netstack.CreateNetTUN([]netip.Addr{deviceRPCIP}, []netip.Addr{}, deviceMTU); check(err) {
		startupErrors <- err
		return
	}

	dev = device.NewDevice(tunnel, conn.NewDefaultBind(), device.NewLogger(device.LogLevelError, "server "))

	dev.SetPrivateKey(config.key.AsDeviceKey())
	dev.IpcSet("listen_port=" + strconv.Itoa(int(config.listenPort)))

	for _, peer_whitelist := range config.peerWhitelist {

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

	if tcpSock, err = network.ListenTCPAddrPort(netip.AddrPortFrom(deviceRPCIP, deviceRPCPort)); check(err) {
		startupErrors <- err
		return
	}

	go server.Serve(tcpSock)

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
