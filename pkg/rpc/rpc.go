package rpc

import (
	"git-indra.lan/indra-labs/indra"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"github.com/multiformats/go-multiaddr"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/tun/netstack"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"net"
	"net/netip"
	"net/rpc"
	"strconv"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

var (
	DefaultIPAddress = netip.MustParseAddr("127.0.37.1")
)

type RPCConfig struct {
	Key            *RPCPrivateKey
	ListenPort     uint16
	Peer_Whitelist []RPCPublicKey
	IP_Whitelist   []multiaddr.Multiaddr
}

func (conf *RPCConfig) IsEnabled() bool {
	return !conf.Key.IsZero()
}

type RPC struct {
	device  *device.Device
	network *netstack.Net
	tunnel  tun.Device
}

func (r *RPC) Start() error {

	log.I.Ln("starting rpc server")

	r.device.Up()

	var err error
	var listener *gonet.TCPListener

	if listener, err = r.network.ListenTCP(&net.TCPAddr{Port: 80}); check(err) {
		return err
	}

	rpc.HandleHTTP()

	go rpc.Accept(listener)

	return nil
}

func (rpc *RPC) Stop() {

	rpc.device.Close()

}

func New(config *RPCConfig) (*RPC, error) {

	var err error
	var r RPC

	if r.tunnel, r.network, err = netstack.CreateNetTUN([]netip.Addr{DefaultIPAddress}, []netip.Addr{}, 1420); check(err) {
		return nil, err
	}

	r.device = device.NewDevice(r.tunnel, conn.NewDefaultBind(), device.NewLogger(device.LogLevelError, "server "))

	r.device.SetPrivateKey(config.Key.AsDeviceKey())
	r.device.IpcSet("listen_port=" + strconv.Itoa(int(config.ListenPort)))

	for _, peer_whitelist := range config.Peer_Whitelist {

		deviceConf := "" +
			"public_key=" + peer_whitelist.HexString() + "\n" +
			"allowed_ip=" + "127.0.37.2" + "/32\n"

		if err = r.device.IpcSet(deviceConf); check(err) {
			return nil, err
		}
	}

	return &r, nil
}
