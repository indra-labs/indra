package storage

import (
	"context"
	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/viper"
)

var (
	fileName string = "indra.db"
)

var (
	db            *badger.DB
	opts          badger.Options
	startupErrors = make(chan error)
	isReady       = make(chan bool)
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

func Run(ctx context.Context) {

	configure()

	log.I.Ln("running storage")

	var err error

	opts = badger.DefaultOptions(viper.GetString(storeFilePathFlag))
	opts.EncryptionKey = key.Bytes()
	opts.IndexCacheSize = 128 << 20
	opts.WithLoggingLevel(badger.WARNING)

	db, err = badger.Open(opts)

	if err != nil {

		check(err)

		startupErrors <- err
		return
	}

	isReady <- true

	select {
	case <-ctx.Done():
		Shutdown()
	}

	return
}
