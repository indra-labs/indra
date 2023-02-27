package storage

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/viper"
	"sync"
)

var (
	fileName string = "indra.db"
	db       *badger.DB
	opts     badger.Options
)

var (
	running sync.Mutex
)

func Run() {

	if !running.TryLock() {
		return
	}

	configure()

	if !attempt_unlock() {
		isLockedChan <- true
		return
	}

	log.I.Ln("storage is ready")
	isReadyChan <- true
}

func Shutdown() (err error) {

	log.I.Ln("shutting down storage")

	if db == nil {
		return nil
	}

	if db.IsClosed() {
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

func attempt_unlock() bool {

	if noKeyProvided {
		return false
	}

	var err error

	log.I.Ln("attempting to unlock database")

	opts = badger.DefaultOptions(viper.GetString(storeFilePathFlag))
	opts.Logger = nil
	opts.IndexCacheSize = 128 << 20
	opts.EncryptionKey = key.Bytes()

	if db, err = badger.Open(opts); check(err) {
		startupErrors <- err
		return false
	}

	log.I.Ln("successfully unlocked database")
	isUnlockedChan <- true

	return true
}
