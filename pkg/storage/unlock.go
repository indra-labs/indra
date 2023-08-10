package storage

import (
	"git.indra-labs.org/dev/ind/pkg/util/options"
	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/viper"
)

func attempt_unlock() (isUnlocked bool, err error) {

	opts = options.Default(viper.GetString(storeFilePathFlag), key[:])

	if db, err = badger.Open(*opts); err != nil {

		db = nil

		return false, err
	}

	return true, nil
}
