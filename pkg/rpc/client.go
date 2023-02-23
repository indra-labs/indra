package rpc

import (
	"context"
	"github.com/multiformats/go-multiaddr"
	"github.com/tutorialedge/go-grpc-tutorial/chat"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/tun/netstack"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"net/netip"
	"strconv"
)

var (
	DefaultClientIPAddr = netip.MustParseAddr("192.168.4.29")
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
		"allowed_ip=" + deviceRPCIP.String() + "/32\n" +
		"persistent_keepalive_interval=" + strconv.Itoa(int(config.Peer.KeepAliveInterval)) + "\n"

	if err = r.device.IpcSet(deviceConf); check(err) {
		return nil, err
	}

	var conn *grpc.ClientConn

	//conn, err = grpc.Dial(
	//	"unix:///tmp/indra.sock",
	//	grpc.WithBlock(),
	//	grpc.WithTransportCredentials(insecure.NewCredentials()),
	//)

	conn, err = grpc.DialContext(context.Background(),
		deviceRPCIP.String()+":80",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, address string) (net.Conn, error) {
			return r.network.DialContext(ctx, "tcp4", address)
		}))

	c := chat.NewChatServiceClient(conn)

	response, err := c.SayHello(context.Background(), &chat.Message{Body: "Hello From Client!"})

	if err != nil {
		check(err)
	}

	log.I.F(response.Body)

	return &r, nil
}
