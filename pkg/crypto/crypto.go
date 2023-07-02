package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"sync"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/gookit/color"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/b32/based32"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/libp2p/go-libp2p/core/crypto"
)

const (
	// BlindLen is the length of the blinding factor in bytes for a Cloak.
	BlindLen = 3

	// HashLen is the length from the hash of the key with the Blinder appended to
	// the Cloak.
	HashLen = 5

	// CloakLen is the sum of the Blinder truncated hash from blinder and public key.
	CloakLen = BlindLen + HashLen

	// PrvKeyLen is the length of a private key, inherited from secp256k1.
	PrvKeyLen = secp256k1.PrivKeyBytesLen

	// PubKeyLen is the length of the serialized key. It is an ECDSA compressed
	// key, 33 bytes with a 2 or 3 prefix to signify the sign of the key.
	PubKeyLen = secp256k1.PubKeyBytesLenCompressed

	// SigLen is the length of the signatures used in Indra, compact keys that can
	// have the public key extracted from them.
	SigLen = 65
)

var (
	// The key types must satisfy these interfaces for libp2p.
	_, _  crypto.Key     = &Prv{}, &Pub{}
	_     crypto.PubKey  = &Pub{}
	_     crypto.PrivKey = &Prv{}
	fails                = log.E.Chk
	log                  = log2.GetLogger()
)

// Blinder is the bytes concatenated after a key to generate the Cloak hash.
type Blinder [BlindLen]byte

// Hash is the truncated SHA256 hash of the Blinder with a PubBytes.
type Hash [HashLen]byte

// Prv is a secp256k1 private key.
type Prv secp256k1.PrivateKey

// PrvBytes is the binary encoded form of a secp256k1 private key.
type PrvBytes [PrvKeyLen]byte

// PubBytes is the serialised form of a secp256k1 public key.
type PubBytes [PubKeyLen]byte

// CloakedPubKey is the blinded hash of a public key used to conceal a message public
// key from attackers.
type CloakedPubKey [CloakLen]byte

func Cloak(b Blinder, key PubBytes) (c CloakedPubKey) {
	h := sha256.Single(append(b[:], key[:]...))
	copy(c[:BlindLen], b[:BlindLen])
	copy(c[BlindLen:BlindLen+HashLen], h[:HashLen])
	return
}

// DerivePub generates a public key from the prv.Pub.
func DerivePub(prv *Prv) *Pub {
	if prv == nil {
		return nil
	}
	return (*Pub)((*secp256k1.PrivateKey)(prv).PubKey())
}

// Equals checks if two binary encoded public keys are equal.
func (pb PubBytes) Equals(qb PubBytes) bool { return pb == qb }

// KeySet is a fast private key generator that uses two random base private keys
// and combines one with the other repeatedly to generate a new, valid private key,
// using scalar multiplication.
type KeySet struct {
	Mutex           sync.Mutex
	Base, Increment *Prv
}

// GeneratePrvKey generates a secp256k1 private key.
func GeneratePrvKey() (prv *Prv, e error) {
	var p *secp256k1.PrivateKey
	if p, e = secp256k1.GeneratePrivateKey(); fails(e) {
		return
	}
	return (*Prv)(p), e
}

// GetCloak returns a value which a receiver with the private key can identify
// the association of a message with the peer in order to retrieve the private
// key to generate the message cipher.
//
// The three byte blinding factor concatenated in front of the public key
// generates the 5 bytes at the end of the CloakedPubKey code. In this way the
// source public key it relates to is hidden to any who don't have this public
// key, which only the parties know.
func GetCloak(s *Pub) (c CloakedPubKey) {
	var blinder Blinder
	var n int
	var e error
	if n, e = rand.Read(blinder[:]); fails(e) && n != BlindLen {
		panic("no entropy")
	}
	c = Cloak(blinder, s.ToBytes())
	return
}

// Next adds Increment to Base, assigns the new private key to the Base and
// returns the new private key.
func (ks *KeySet) Next() (n *Prv) {
	ks.Mutex.Lock()
	next := ks.Base.Key.Add(&ks.Increment.Key)
	ks.Base.Key = *next
	n = &Prv{Key: *next}
	ks.Mutex.Unlock()
	return
}

// Next2 returns two private keys from the KeySet.
func (ks *KeySet) Next2() (n [2]*Prv) {
	for i := range n {
		n[i] = ks.Next()
	}
	return
}

// Next3 returns three private keys from the KeySet.
func (ks *KeySet) Next3() (n [3]*Prv) {
	for i := range n {
		n[i] = ks.Next()
	}
	return
}

// Match uses the cached public key and the provided blinding factor to match the
// source public key so the packet address field is only recognisable to the
// intended recipient.
func Match(r CloakedPubKey, k PubBytes) bool {
	var b Blinder
	copy(b[:], r[:BlindLen])
	hash := Cloak(b, k)
	return r == hash
}

// PrvFromBased32 decodes a Based32 encoded private key.
func PrvFromBased32(s string) (k *Prv, e error) {
	ss := []byte(s)
	var b slice.Bytes
	b, e = based32.Codec.Decode("a" + string(ss))
	k = PrvKeyFromBytes(b)
	return
}

// PrvKeyFromBytes converts a byte slice into a private key.
func PrvKeyFromBytes(b []byte) *Prv {
	return (*Prv)(secp256k1.PrivKeyFromBytes(b))
}

// ToBased32 returns the Based32 encoded string of the private key.
func (p *Prv) ToBased32() (s string) {
	b := p.ToBytes()
	var e error
	if s, e = based32.Codec.Encode(b[:]); fails(e) {
	}
	ss := []byte(s[1:])
	return string(ss)
}

// ToBytes returns the Bytes serialized form.
func (p *Prv) ToBytes() (b PrvBytes) {
	br := (*secp256k1.PrivateKey)(p).Serialize()
	copy(b[:], br[:PrvKeyLen])
	return
}

// PubFromBytes converts a byte slice into a public key, if it is valid and on
// the secp256k1 elliptic curve.
func PubFromBytes(b []byte) (pub *Pub, e error) {
	var p *secp256k1.PublicKey
	if p, e = secp256k1.ParsePubKey(b); fails(e) {
		return
	}
	pub = (*Pub)(p)
	return
}

// ToPublicKey unwraps the secp256k1.PublicKey inside the Pub.
func (k *Pub) ToPublicKey() *secp256k1.PublicKey {
	return (*secp256k1.PublicKey)(k)
}

// Recover the public key corresponding to the signing private key used to
// create a signature on the hash of a message.
func (sb SigBytes) Recover(hash sha256.Hash) (p *Pub, e error) {
	var pk *secp256k1.PublicKey
	// We are only using compressed keys, so we can ignore the compressed bool.
	if pk, _, e = ecdsa.RecoverCompact(sb[:], hash[:]); !fails(e) {
		p = (*Pub)(pk)
	}
	return
}

// SigBytes is an ECDSA BIP62 formatted compact signature which allows the
// recovery of the public key from the signature.
type SigBytes [SigLen]byte

// NewSigner creates a new KeySet which enables (relatively) fast generation of
// new private keys for use with ECDH encryption by using scalar addition.
func NewSigner() (first *Prv, ks *KeySet, e error) {
	ks = &KeySet{}
	if ks.Base, e = GeneratePrvKey(); fails(e) {
		return
	}
	if ks.Increment, e = GeneratePrvKey(); fails(e) {
		return
	}
	first = ks.Base
	return
}

// Zero out a private key to prevent key scraping from memory.
func (p *Prv) Zero() { (*secp256k1.PrivateKey)(p).Zero() }

// Pub is a public key.
type Pub secp256k1.PublicKey

// PubFromBased32 decodes a Based32 encoded form of the Pub.
func PubFromBased32(s string) (k *Pub, e error) {
	ss := []byte(s)
	var b slice.Bytes
	b, e = based32.Codec.Decode("ayb" + string(ss))
	return PubFromBytes(b)
}

// String returns a Based32 encoded string of the public key.
func (k *Pub) String() (s string) { return k.ToBased32() }

// ToBased32 returns the Based32 formatted form of the public key.
func (k *Pub) ToBased32() (s string) {
	b := k.ToBytes()
	var e error
	if s, e = based32.Codec.Encode(b[:]); fails(e) {
	}
	ss := []byte(s)[3:]
	return string(ss)
}

// ToBased32Abbreviated returns a concatenated form with the middle replaced with an ellipsis.
func (k *Pub) ToBased32Abbreviated() (s string) {
	s = k.ToBased32()
	s = s[:13] + "..." + s[len(s)-8:]
	return color.LightGreen.Sprint(string(s))
}

// ToBytes returns the compressed 33 byte form of the pubkey as used in wire and
// storage forms.
func (k *Pub) ToBytes() (p PubBytes) {
	b := (*secp256k1.PublicKey)(k).SerializeCompressed()
	copy(p[:], b)
	return
}

// ToHex returns the hex encoding of a Pub.
func (k *Pub) ToHex() (s string, e error) {
	b := k.ToBytes()
	s = hex.EncodeToString(b[:])
	return
}

// SigFromBased32 decodes a SigBytes encoded in Based32.
func SigFromBased32(s string) (sig SigBytes, e error) {
	var ss slice.Bytes
	ss, e = based32.Codec.Decode("aq" + s)
	copy(sig[:], ss)
	return
}

// Sign produces an ECDSA BIP62 compact signature.
func Sign(prv *Prv, hash sha256.Hash) (sig SigBytes, e error) {
	copy(sig[:],
		ecdsa.SignCompact((*secp256k1.PrivateKey)(prv), hash[:], true))
	return
}

// String returns the Based32 encoded form of a signature.
func (sb SigBytes) String() string {
	o, _ := based32.Codec.Encode(sb[:])
	return o[2:]
}

// String returns the based32 form of a public key's compact encoded bytes.
func (pb PubBytes) String() (s string) { return pb.ToBased32() }

// ToBased32 returns the based32 form of a public key's compact encoded bytes.
func (pb PubBytes) ToBased32() (s string) {
	var e error
	if s, e = based32.Codec.Encode(pb[:]); fails(e) {
	}
	ss := []byte(s)[3:]
	return string(ss)

}

// Zero zeroes out a private key in serial form.
func (pb PrvBytes) Zero() { copy(pb[:], zeroPrv()) }

func zeroPrv() []byte {
	z := PrvBytes{}
	return z[:]
}
