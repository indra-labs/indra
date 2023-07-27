package options

import "github.com/dgraph-io/badger/v3"

// Default returns a pointer to badger.Options to be used to open
// Indra's main data store.
//
// This is separated from the seed's usage of it in order to make test data
// stores without duplicating this common configuration setting.
func Default(filePath string, key []byte) *badger.Options {

	o := badger.DefaultOptions(filePath)

	// If log level is above info maybe we do want this enabled?
	o.Logger = nil

	// This works out as 1 << 27, ie 256kb. Should it be 1<<30 1Mb?
	o.IndexCacheSize = 128 << 20

	o.EncryptionKey = key[:]

	return &o
}
