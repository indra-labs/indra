package ecdh

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

// Keysize is the size of both Privkey and Pubkey for ECDH
const Keysize = 32

// Privkey is a private key for use with ECDH
type Privkey [Keysize]byte

// Pubkey is a public key generated via curve25519.ScalarBaseMult
type Pubkey [Keysize]byte

// Keypair is pointers to a private and public key bundled into a struct
type Keypair struct {
	*Privkey
	*Pubkey
}

// KeySizeError indicates a provided slice of bytes is the wrong length
var KeySizeError error = fmt.Errorf("key length must be precisely %d bytes",
	Keysize)

func GeneratePrivkey() (prv *Privkey, err error) {
	var pk [32]byte
	_, err = rand.Read(pk[:])
	if log.E.Chk(err) {
		return
	}
	pk[0] &= 248
	pk[31] &= 127
	pk[31] |= 64
	pr := Privkey(pk)
	prv = &pr
	return
}

func (prv Privkey) GeneratePubkey() (pub *Pubkey) {
	var dst [32]byte
	curve25519.ScalarBaseMult(&dst, (*[32]byte)(&prv))
	pu := Pubkey(dst)
	pub = &pu
	return
}

func (prv Privkey) ToBytes() (b []byte) {
	b = make([]byte, Keysize)
	copy(b, prv[:])
	return
}

func (pub Pubkey) ToBytes() (b []byte) {
	b = make([]byte, Keysize)
	copy(b, pub[:])
	return
}

func PrivkeyFromBytes(b []byte) (prv *Privkey, err error) {
	if len(b) != Keysize {
		err = KeySizeError
		return
	}
	var k [Keysize]byte
	copy(k[:], b)
	prv = (*Privkey)(&k)
	return
}

func PubkeyFromBytes(b []byte) (pub *Pubkey, err error) {
	if len(b) != Keysize {
		err = KeySizeError
		return
	}
	var k [Keysize]byte
	copy(k[:], b)
	pub = (*Pubkey)(&k)
	return
}

// GenerateKeypair generates a new curve25519 private/public key pair.
func GenerateKeypair() (keypair Keypair, err error) {
	if keypair.Privkey, err = GeneratePrivkey(); log.E.Chk(err) {
		return
	}
	keypair.Pubkey = keypair.Privkey.GeneratePubkey()
	return
}

// KeypairFromPrivkeyBytes generates a Keypair from raw bytes
func KeypairFromPrivkeyBytes(b []byte) (keypair *Keypair, err error) {
	keypair = &Keypair{}
	if keypair.Privkey, err = PrivkeyFromBytes(b); log.E.Chk(err) {
		return
	}
	keypair.Pubkey = keypair.Privkey.GeneratePubkey()
	return
}

// ComputeSecret uses a Keypair and a Pubkey to generate a secret to use in
// encryption. This formula generates the same value from one private to another
// public key as the public key of the first with the private key from the
// second. Thus neither party must send a secret over the network.
func (kp Privkey) ComputeSecret(p *Pubkey) (sec []byte, err error) {
	pub := [Keysize]byte(*p)
	if sec, err = curve25519.X25519(kp[:], pub[:]); log.E.Chk(err) {
	}
	return
}
