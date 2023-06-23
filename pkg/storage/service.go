package storage

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/indra-labs/indra/pkg/rpc"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
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

func run() {

	if noKeyProvided {
		log.I.Ln("storage is locked")
		isLockedChan <- true
		return
	}

	log.I.Ln("attempting to unlock database")
	isUnlocked, err := attempt_unlock()

	if !isUnlocked {
		log.I.Ln("unlock attempt failed")
		startupErrors <- err
	}

	log.I.Ln("successfully unlocked database")
	isUnlockedChan <- true
}

func Run() {

	if !running.TryLock() {
		return
	}

	configure()

	run()

signals:
	for {
		select {
		case err := <-rpc.WhenStartFailed():
			startupErrors <- err
			return
		case <-WhenIsUnlocked():
			rpc.Shutdown()
			break signals
		case <-WhenIsLocked():

			log.I.Ln("running rpc server, with unlock service")

			go rpc.RunWith(
				func(srv *grpc.Server) {
					RegisterUnlockServiceServer(srv, NewUnlockService())
				},
				rpc.WithUnixPath(viper.GetString(rpc.UnixPathFlag)),
			)
		case <-rpc.IsReady():
			log.I.Ln("... awaiting unlock over rpc")
		}
	}

	log.I.Ln("running garbage collection before ready")
	db.RunValueLogGC(0.5)

	log.I.Ln("storage is ready")
	isReadyChan <- true
}

func Shutdown() (err error) {

	rpc.Shutdown()

	log.I.Ln("shutting down storage")

	if db == nil {
		log.I.Ln("- storage was never started")
		return nil
	}

	log.I.Ln("- storage db closing, it may take a minute...")

	db.RunValueLogGC(0.5)

	if err = db.Close(); err != nil {
		log.W.Ln("- storage shutdown warning: ", err)
	}

	log.I.Ln("- storage shutdown completed")

	return
}

func Txn(tx func(txn *badger.Txn) error, update bool) (err error) {

	txn := db.NewTransaction(update)

	if err = tx(txn); err != nil {
		return
	}

	return txn.Commit()
}

func View(fn func(txn *badger.Txn) error) error {
	return db.View(fn)
}

func Update(fn func(txn *badger.Txn) error) error {
	return db.Update(fn)
}

func DB() *badger.DB {
	return db
}
