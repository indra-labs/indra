package examples

import (
	"context"
	"os"
	
	"github.com/tutorialedge/go-grpc-tutorial/chat"
	"google.golang.org/grpc"
	
	"git-indra.lan/indra-labs/indra/pkg/rpc"
)

func UnixHello(ctx context.Context) {
	
	var err error
	var conn *grpc.ClientConn
	
	conn, err = rpc.Dial("unix:///tmp/indra.sock")
	
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
