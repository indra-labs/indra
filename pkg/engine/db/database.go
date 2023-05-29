package db

import (
	badger "github.com/ipfs/go-ds-badger2"

	"github.com/indra-labs/indra"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

type Database struct {
	*badger.Datastore
}

func New(path string) (db *Database, e error) {
	var ds *badger.Datastore
	ds, e = badger.NewDatastore(path, nil)
	if fails(e) {
		return
	}
	return &Database{ds}, nil
}
