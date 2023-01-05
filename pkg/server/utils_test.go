package server

import (
	"crypto/rand"
	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p/core/crypto"
	"testing"
)

func TestBase58(t *testing.T){

	var err error
	var priv1, priv2 crypto.PrivKey
	var keyStr1, keyStr2 string

	// Generate priv
	priv1, _, err = crypto.GenerateSecp256k1Key(rand.Reader)

	if keyStr1, err = Base58Encode(priv1); err != nil {
		t.Error("base58encode error: ", err)
	}

	spew.Dump(priv1)

	if priv2, err = Base58Decode(keyStr1); err != nil {
		t.Error("base58decode error: ", err)
	}

	if !priv1.Equals(priv2) {
		t.Error("Keys are not equal!")
	}

	if keyStr2, err = Base58Encode(priv2); err != nil {
		t.Error("base58encode error: ", err)
	}

	spew.Dump(priv2)

	if keyStr1 != keyStr2 {
		t.Error("Keys are not equal!")
	}
}
