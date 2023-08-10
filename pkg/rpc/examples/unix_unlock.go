package examples

import (
	"context"
	"git.indra-labs.org/dev/ind/pkg/rpc"
	"git.indra-labs.org/dev/ind/pkg/storage"
	"google.golang.org/grpc"
	"os"
)

func UnixUnlock(ctx context.Context) {

	var err error
	var conn *grpc.ClientConn

	conn, err = rpc.Dial("unix:///tmp/indra.sock")

	if err != nil {
		check(err)
		os.Exit(1)
	}

	u := storage.NewUnlockServiceClient(conn)

	_, err = u.Unlock(ctx, &storage.UnlockRequest{
		Key: "979nrx9ry9Re6UqWXYaGqLEne8NS7TzgHFiS8KARABV8",
	})

	if err != nil {
		check(err)
		return
	}
}
