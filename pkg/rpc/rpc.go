package rpc

import (
	"git-indra.lan/indra-labs/indra"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"github.com/davecgh/go-spew/spew"
	"github.com/multiformats/go-multiaddr"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/tun/netstack"
	"net/netip"
	"strconv"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
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
	device *device.Device
}

func (rpc *RPC) Start() {
	rpc.device.Up()
}

func (rpc *RPC) Stop() {

	rpc.device.Close()
}

func New(config *RPCConfig) (*RPC, error) {

	var err error
	var r RPC

	var tunnel tun.Device
	//var network *netstack.Net

	if tunnel, _, err = netstack.CreateNetTUN([]netip.Addr{}, []netip.Addr{}, 1420); check(err) {
		return nil, err
	}

	r.device = device.NewDevice(tunnel, conn.NewDefaultBind(), device.NewLogger(device.LogLevelVerbose, ""))

	r.device.SetPrivateKey(config.Key.AsDeviceKey())
	r.device.IpcSet("listen_port=" + strconv.Itoa(int(config.ListenPort)))

	var peer *device.Peer

	for _, peer_whitelist := range config.Peer_Whitelist {

		if peer, err = r.device.NewPeer(peer_whitelist.AsDeviceKey()); check(err) {
			return nil, err
		}

		spew.Dump(peer.String())
		peer.Start()

	}

	return &r, nil
}
