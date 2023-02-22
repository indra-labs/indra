package rpc

import (
	"github.com/multiformats/go-multiaddr"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/tun/netstack"
	"net/netip"
	"strconv"
)

var (
	DefaultClientIPAddr = netip.MustParseAddr("127.0.37.2")
)

type Peer struct {
	Endpoint          multiaddr.Multiaddr
	PublicKey         *RPCPublicKey
	PreSharedKey      RPCPrivateKey
	KeepAliveInterval uint8
}

type ClientConfig struct {
	Key  RPCPrivateKey
	Peer *Peer
}

var (
	DefaultClientConfig = &ClientConfig{
		Key: DecodePrivateKey("Aj9CfbE1pXEVxPfjSaTwdY3B4kYHbwsTSyT3nrc34ATN"),
		Peer: &Peer{
			Endpoint:          multiaddr.StringCast("/ip4/127.0.0.1/udp/18222"),
			PublicKey:         DecodePublicKey("G52UmsQpUmN2zFMkJaP9rwCvqQJzi1yHKA9RTrLJTk9f"),
			KeepAliveInterval: 5,
		},
	}
)

type RPCClient struct {
	device  *device.Device
	network *netstack.Net
}

func (r *RPCClient) Start() {
	r.device.Up()
}

func (rpc *RPCClient) Stop() {
	rpc.device.Close()
}

func NewClient(config *ClientConfig) (*RPCClient, error) {

	var err error
	var r RPCClient

	var tunnel tun.Device

	if tunnel, r.network, err = netstack.CreateNetTUN([]netip.Addr{DefaultClientIPAddr}, []netip.Addr{}, 1420); check(err) {
		return nil, err
	}

	r.device = device.NewDevice(tunnel, conn.NewDefaultBind(), device.NewLogger(device.LogLevelError, "client "))

	r.device.SetPrivateKey(config.Key.AsDeviceKey())

	deviceConf := "" +
		"public_key=" + config.Peer.PublicKey.HexString() + "\n" +
		"endpoint=0.0.0.0:18222" + "\n" +
		"allowed_ip=" + deviceIP.String() + "/32\n" +
		"persistent_keepalive_interval=" + strconv.Itoa(int(config.Peer.KeepAliveInterval)) + "\n"

	if err = r.device.IpcSet(deviceConf); check(err) {
		return nil, err
	}

	return &r, nil
}
