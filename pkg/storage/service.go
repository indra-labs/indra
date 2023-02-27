package storage

import (
	"context"
	"git-indra.lan/indra-labs/indra/pkg/interrupt"
	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/viper"
	"sync"
)

var (
	fileName string = "indra.db"
)

var (
	db            *badger.DB
	opts          badger.Options
	startupErrors = make(chan error, 128)
	isLockedChan  = make(chan bool, 1)
	isReadyChan   = make(chan bool, 1)
)

func CantStart() chan error {
	return startupErrors
}

func IsLocked() chan bool {
	return isLockedChan
}

func IsReady() chan bool {
	return isReadyChan
}

func Shutdown() (err error) {

	log.I.Ln("shutting down storage")

	if db == nil {
		return nil
	}

	if err = db.Close(); check(err) {
		return
	}

	log.I.Ln("storage shutdown completed")

	return
}

func Txn(tx func(txn *badger.Txn) error, update bool) error {

	txn := db.NewTransaction(update)

	return tx(txn)
}

var (
	running sync.Mutex
)

func open() {

	var err error

	opts.EncryptionKey = key.Bytes()

	if db, err = badger.Open(opts); check(err) {
		startupErrors <- err
		return
	}

	log.I.Ln("successfully opened database")
	log.I.Ln("storage is ready")

	isReadyChan <- true
}

func Run(ctx context.Context) {

	if !running.TryLock() {
		return
	}

	configure()

	opts = badger.DefaultOptions(viper.GetString(storeFilePathFlag))
	opts.IndexCacheSize = 128 << 20
	opts.Logger = nil

	if !isLocked {

		log.I.Ln("attempting to open database with key")

		open()
	}

	isLockedChan <- true

	lockedCtx, cancel := context.WithCancel(context.Background())

	interrupt.AddHandler(cancel)

	for {
		select {
		case <-IsReady():
			log.I.Ln("storage is ready")

		//case <-unlock.IsSuccessful():
		//
		//	log.I.Ln("storage successfully unlocked")
		//
		//	isReadyChan <- true

		case <-lockedCtx.Done():
			Shutdown()
			return
		}
	}
}
