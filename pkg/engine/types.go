package engine

import (
	"crypto/cipher"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type Ciphers [3]sha256.Hash
type Nonces [3]nonce.IV
type Privs [3]*crypto.Prv
type Pubs [3]*crypto.Pub
type Keys struct {
	Pub   *crypto.Pub
	Bytes crypto.PubBytes
	Prv   *crypto.Prv
}

func GenerateKeys() (k *Keys, e error) {
	k = &Keys{}
	if k.Prv, e = crypto.GeneratePrvKey(); fails(e) {
		return
	}
	k.Pub = crypto.DerivePub(k.Prv)
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

func MakeKeys(pr *crypto.Prv) *Keys {
	pubkey := crypto.DerivePub(pr)
	return &Keys{pubkey, pubkey.ToBytes(), pr}
}

func Encipher(b slice.Bytes, iv nonce.IV, from *crypto.Prv, to *crypto.Pub,
	note string) (e error) {
	
	var blk cipher.Block
	if blk = ciph.GetBlock(from, to, note); fails(e) {
		return
	}
	ciph.Encipher(blk, iv, b)
	return
}
