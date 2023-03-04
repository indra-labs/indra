package rpc

import (
	"errors"
	"github.com/dgraph-io/badger/v3"
)

var (
	ErrKeyNotExists error = errors.New("key not found")
)

var (
	tunKeyKey = "rpc-tun-key"
)

type Store interface {
	Reset()

	SetKey(key *PrivateKey) error

	GetKey() (*PrivateKey, error)
}

type storeMem struct {
	key *PrivateKey
}

func (s *storeMem) Reset() {
	s.key = nil
}

func (s *storeMem) SetKey(key *PrivateKey) error {
	s.key = key
	return nil
}

func (s *storeMem) GetKey() (*PrivateKey, error) {

	if s.key == nil {
		return nil, ErrKeyNotExists
	}

	if s.key.IsZero() {
		return nil, ErrKeyNotExists
	}

	return s.key, nil
}

type BadgerStore struct {
	*badger.DB
}

func (s *BadgerStore) Reset() {

	s.Update(func(txn *badger.Txn) error {
		txn.Delete([]byte(tunKeyKey))

		return nil
	})
}

func (s *BadgerStore) SetKey(key *PrivateKey) error {

	s.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(tunKeyKey), key.Bytes())
		return err
	})

	return nil
}

func (s *BadgerStore) GetKey() (*PrivateKey, error) {

	var err error
	var item *badger.Item

	err = s.View(func(txn *badger.Txn) error {
		item, err = txn.Get([]byte(tunKeyKey))
		return err
	})

	if err == badger.ErrKeyNotFound {
		return nil, ErrKeyNotExists
	}

	var key PrivateKey

	err = item.Value(func(val []byte) error {
		key.DecodeBytes(val)
		return nil
	})

	return &key, err
}
