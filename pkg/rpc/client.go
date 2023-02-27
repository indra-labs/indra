package rpc

import (
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/tun/netstack"
	"net/netip"
	"strconv"
)

func getNetworkInstance(opts *dialOptions) (net *netstack.Net, err error) {

	var tunnel tun.Device

	if tunnel, net, err = netstack.CreateNetTUN([]netip.Addr{netip.MustParseAddr(opts.peerRPCIP)}, []netip.Addr{}, opts.mtu); check(err) {
		return nil, err
	}

	dev := device.NewDevice(tunnel, conn.NewDefaultBind(), device.NewLogger(device.LogLevelError, "client "))

	dev.SetPrivateKey(opts.key.AsDeviceKey())

	deviceConf := "" +
		"public_key=" + opts.peerPubKey.HexString() + "\n" +
		"endpoint=" + opts.endpoint.String() + "\n" +
		"allowed_ip=" + opts.rpcEndpoint.Address() + "/32\n" +
		"persistent_keepalive_interval=" + strconv.Itoa(opts.keepAliveInterval) + "\n"

	if err = dev.IpcSet(deviceConf); check(err) {
		return
	}

	return net, nil
}
