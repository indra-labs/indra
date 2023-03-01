package storage

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/viper"
)

func attempt_unlock() (isUnlocked bool, err error) {

	opts = badger.DefaultOptions(viper.GetString(storeFilePathFlag))
	opts.Logger = nil
	opts.IndexCacheSize = 128 << 20
	opts.EncryptionKey = key.Bytes()

	if db, err = badger.Open(opts); err != nil {

		db = nil

		return false, err
	}

	return true, nil
}
