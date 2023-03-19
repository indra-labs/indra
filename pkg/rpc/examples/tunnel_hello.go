package examples

import (
	"context"
	"os"
	
	"github.com/tutorialedge/go-grpc-tutorial/chat"
	"google.golang.org/grpc"
	
	"git-indra.lan/indra-labs/indra/pkg/rpc"
)

func TunnelHello(ctx context.Context) {
	
	var err error
	var conn *grpc.ClientConn
	
	conn, err = rpc.DialContext(ctx,
		"noise://[::1]:18222",
		rpc.WithPrivateKey("Aj9CfbE1pXEVxPfjSaTwdY3B4kYHbwsTSyT3nrc34ATN"),
		rpc.WithPeer("G52UmsQpUmN2zFMkJaP9rwCvqQJzi1yHKA9RTrLJTk9f"),
		rpc.WithKeepAliveInterval(5),
	)
	
	if err != nil {
		check(err)
		os.Exit(1)
	}
	
	c := chat.NewChatServiceClient(conn)
	
	response, err := c.SayHello(context.Background(), &chat.Message{Body: "Hello From Alice!"})
	
	if err != nil {
		check(err)
		return
	}
	
	log.I.F(response.Body)
}
