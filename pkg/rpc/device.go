package rpc

import (
	"golang.zx2c4.com/wireguard/device"
	"net/netip"
	"strconv"
)

var (
	dev           *device.Device
	deviceRPCIP   netip.Addr = netip.MustParseAddr("192.168.37.1")
	deviceRPCPort uint16     = 80
)

func configureDevice() {

	var err error

	dev.SetPrivateKey(tunKey.AsDeviceKey())
	dev.IpcSet("listen_port=" + strconv.Itoa(int(tunnelPort)))

	for _, peer_whitelist := range tunWhitelist {

		deviceConf := "" +
			"public_key=" + peer_whitelist.HexString() + "\n" +
			"allowed_ip=192.168.37.2/32\n"

		if err = dev.IpcSet(deviceConf); check(err) {
			startupErrors <- err
		}
	}
}
