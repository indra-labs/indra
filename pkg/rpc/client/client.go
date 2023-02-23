package client

import (
	"context"
	"github.com/tutorialedge/go-grpc-tutorial/chat"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/tun/netstack"
	"google.golang.org/grpc"
	"net/netip"
	"os"
	"strconv"
)

var (
	tunnel  tun.Device
	network *netstack.Net
	dev     *device.Device
)

func getNetworkInstance(opts *dialOptions) (err error) {

	if tunnel, network, err = netstack.CreateNetTUN([]netip.Addr{netip.MustParseAddr(opts.peerRPCIP)}, []netip.Addr{}, 1420); check(err) {
		return
	}

	dev = device.NewDevice(tunnel, conn.NewDefaultBind(), device.NewLogger(device.LogLevelError, "client "))

	dev.SetPrivateKey(opts.key.AsDeviceKey())

	deviceConf := "" +
		"public_key=" + opts.peerPubKey.HexString() + "\n" +
		"endpoint=0.0.0.0:18222" + "\n" +
		"allowed_ip=" + rpcEndpointIp + "/32\n" +
		"persistent_keepalive_interval=" + strconv.Itoa(opts.keepAliveInterval) + "\n"

	if err = dev.IpcSet(deviceConf); check(err) {
		return
	}

	return nil
}

func Run(ctx context.Context) {

	var err error
	var conn *grpc.ClientConn

	conn, err = Dial("unix:///tmp/indra.sock")

	//conn, err = DialContext(ctx,
	//	"noise://0.0.0.0:18222",
	//	WithPrivateKey("Aj9CfbE1pXEVxPfjSaTwdY3B4kYHbwsTSyT3nrc34ATN"),
	//	WithPeer("G52UmsQpUmN2zFMkJaP9rwCvqQJzi1yHKA9RTrLJTk9f"),
	//	WithKeepAliveInterval(5),
	//)

	if err != nil {
		check(err)
		os.Exit(1)
	}

	c := chat.NewChatServiceClient(conn)

	response, err := c.SayHello(context.Background(), &chat.Message{Body: "Hello From Client!"})

	if err != nil {
		check(err)
	}

	log.I.F(response.Body)
}
