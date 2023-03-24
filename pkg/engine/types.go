package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
)

type Ciphers [3]sha256.Hash
type Nonces [3]nonce.IV
type Privs [3]*prv.Key
type Pubs [3]*pub.Key
type Keys struct {
	Pub   *pub.Key
	Bytes pub.Bytes
	Prv   *prv.Key
}

func MakeKeys(pr *prv.Key) *Keys {
	pubkey := pub.Derive(pr)
	return &Keys{pubkey, pubkey.ToBytes(), pr}
}
