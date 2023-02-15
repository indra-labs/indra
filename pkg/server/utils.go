package server

import (
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/libp2p/go-libp2p/core/crypto"
)

var hnd = "ind"

func bech32encode(key crypto.PrivKey) (keyStr string, err error) {

	var raw []byte

	if raw, err = key.Raw(); check(err) {
		return
	}

	var conv []byte

	if conv, err = bech32.ConvertBits(raw, 8, 5, true); check(err) {
		return
	}

	if keyStr, err = bech32.Encode("ind", conv); check(err) {
		return
	}

	return
}

func bech32decode(keyStr string) (privKey crypto.PrivKey, err error) {

	// var hnd string
	var key []byte

	if _, key, err = bech32.Decode(keyStr); check(err) {
		return
	}

	if privKey, err = crypto.UnmarshalSecp256k1PrivateKey(key); check(err) {
		return
	}

	return privKey, nil
}

func GeneratePrivKey() (privKey crypto.PrivKey) {

	var err error

	if privKey, _, err = crypto.GenerateKeyPair(crypto.Secp256k1, 0); check(err) {
		return
	}

	return
}

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

func GetOrGeneratePrivKey(key string) (privKey crypto.PrivKey, err error) {

	//if key == "" {
	//
	//	privKey = server.GeneratePrivKey()
	//
	//	if key, err = server.Base58Encode(privKey); check(err) {
	//		return
	//	}
	//}
	//
	//if privKey, err = server.Base58Decode(key); check(err) {
	//	return
	//}
	//
	//spew.Dump(key)
	//spew.Dump(privKey)
	//
	//server.DefaultConfig.PrivKey = privKey

	return nil, nil
}
