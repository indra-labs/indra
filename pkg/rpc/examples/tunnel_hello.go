package examples

import (
	"context"
	"git-indra.lan/indra-labs/indra/pkg/rpc"
	"github.com/tutorialedge/go-grpc-tutorial/chat"
	"google.golang.org/grpc"
	"os"
)

func TunnelHello(ctx context.Context) {

	var err error
	var conn *grpc.ClientConn

	//conn, err = Dial("unix:///tmp/indra.sock")

	conn, err = rpc.DialContext(ctx,
		"noise://0.0.0.0:18222",
		rpc.WithPrivateKey("Aj9CfbE1pXEVxPfjSaTwdY3B4kYHbwsTSyT3nrc34ATN"),
		rpc.WithPeer("G52UmsQpUmN2zFMkJaP9rwCvqQJzi1yHKA9RTrLJTk9f"),
		rpc.WithKeepAliveInterval(5),
	)

	if err != nil {
		check(err)
		os.Exit(1)
	}

	c := chat.NewChatServiceClient(conn)

	response, err := c.SayHello(context.Background(), &chat.Message{Body: "Hello From Client!"})

	if err != nil {
		check(err)
		return
	}

	log.I.F(response.Body)
}
