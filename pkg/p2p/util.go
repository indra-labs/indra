package p2p

import (
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/libp2p/go-libp2p/core/crypto"
)

func Base58Encode(priv crypto.PrivKey) (key string, err error) {

	var raw []byte

	raw, err = priv.Raw()

	key = base58.Encode(raw)

	return
}

func Base58Decode(key string) (priv crypto.PrivKey, err error) {

	var raw []byte

	raw = base58.Decode(key)

	if priv, _ = crypto.UnmarshalSecp256k1PrivateKey(raw); check(err) {
		return
	}

	return
}
