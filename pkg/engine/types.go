package engine

import (
	"crypto/cipher"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
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

func GenerateKeys() (k *Keys, e error) {
	k = &Keys{}
	if k.Prv, e = prv.GenerateKey(); fails(e) {
		return
	}
	k.Pub = pub.Derive(k.Prv)
	k.Bytes = k.Pub.ToBytes()
	return
}

func Generate2Keys() (one, two *Keys, e error) {
	if one, e = GenerateKeys(); fails(e) {
		return
	}
	if two, e = GenerateKeys(); fails(e) {
		return
	}
	return
}

func MakeKeys(pr *prv.Key) *Keys {
	pubkey := pub.Derive(pr)
	return &Keys{pubkey, pubkey.ToBytes(), pr}
}

func Encipher(b slice.Bytes, iv nonce.IV, from *prv.Key, to *pub.Key,
	note string) (e error) {
	
	var blk cipher.Block
	if blk = ciph.GetBlock(from, to, note); fails(e) {
		return
	}
	ciph.Encipher(blk, iv, b)
	return
}
