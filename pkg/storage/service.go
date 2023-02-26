package storage

import (
	"context"
	"git-indra.lan/indra-labs/indra/pkg/rpc"
	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"sync"
)

var (
	fileName string = "indra.db"
)

var (
	db            *badger.DB
	opts          badger.Options
	startupErrors = make(chan error, 128)
	isReady       = make(chan bool, 1)
)

func CantStart() chan error {
	return startupErrors
}

func IsReady() chan bool {
	return isReady
}

func Shutdown() (err error) {

	log.I.Ln("shutting down storage")

	if db == nil {
		return nil
	}

	if err = db.Close(); check(err) {
		return
	}

	log.I.Ln("storage shutdown complete")

	return
}

func Txn(tx func(txn *badger.Txn) error, update bool) error {

	txn := db.NewTransaction(update)

	return tx(txn)
}

var (
	running sync.Mutex
)

func Run(ctx context.Context) {

	if !running.TryLock() {
		return
	}

	configure()

	opts = badger.DefaultOptions(viper.GetString(storeFilePathFlag))
	opts.IndexCacheSize = 128 << 20
	opts.Logger = nil

	if isRPCUnlockable {

		var unlockService = NewUnlockService()

		go rpc.RunWith(ctx, func(srv *grpc.Server) {
			RegisterUnlockServiceServer(srv, unlockService)
		})

		select {
		case <-IsReady():
			return
		case <-ctx.Done():
			rpc.Shutdown(context.Background())
			return
		}
	}

	opts.EncryptionKey = key.Bytes()

	log.I.Ln("running storage")

	isReady <- true

	select {
	case <-ctx.Done():
		Shutdown()
	}

	return
}
