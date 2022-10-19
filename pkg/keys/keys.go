package keys

import (
	"github.com/Indra-Labs/indra/pkg/fing"
	"github.com/Indra-Labs/indra/pkg/hasj"
	"github.com/Indra-Labs/indra/pkg/keys/schnorr"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

const (
	PrivkeyLen = secp256k1.PrivKeyBytesLen
	PubkeyLen  = schnorr.PubKeyBytesLen
)

// Privkey is a private key.
type Privkey secp256k1.PrivateKey
type PrivkeyBytes [PrivkeyLen]byte

// Pubkey is a public key.
type Pubkey secp256k1.PublicKey
type PubkeyBytes [PubkeyLen]byte

// GeneratePrivkey generates a private key.
func GeneratePrivkey() (prv *Privkey, e error) {
	var p *secp256k1.PrivateKey
	if p, e = secp256k1.GeneratePrivateKey(); log.I.Chk(e) {
		return
	}
	prv = (*Privkey)(p)
	return
}

// Zero zeroes out a private key to prevent key scraping from memory.
func (prv *Privkey) Zero() {
	(*secp256k1.PrivateKey)(prv).Zero()
}

// Pubkey generates a public key from the Privkey.
func (prv *Privkey) Pubkey() *Pubkey {
	return (*Pubkey)((*secp256k1.PrivateKey)(prv).PubKey())
}

// PubkeyFromBytes converts a byte slice into a public key, if it is valid.
func PubkeyFromBytes(b []byte) (pub *Pubkey, e error) {
	var p *secp256k1.PublicKey
	if p, e = secp256k1.ParsePubKey(b); log.E.Chk(e) {
		return
	}
	pub = (*Pubkey)(p)
	return
}

// Fingerprint generates a fingerprint from a Pubkey.
func (pub Pubkey) Fingerprint() (fp fing.Fingerprint) {
	h := sha256.Double(pub.Serialize()[:])
	copy(fp[:], h[:fing.FingerprintLen])
	return
}

// Fingerprint generates a fingerprint from a PubkeyBytes.
func (pb PubkeyBytes) Fingerprint() (fp *fing.Fingerprint) {
	fp = &fing.Fingerprint{}
	h := sha256.Double(pb[:])
	copy(fp[:], h[:fing.FingerprintLen])
	return
}

// Serialize returns the PrivkeyBytes serialized form.
func (prv *Privkey) Serialize() (b *PrivkeyBytes) {
	b = &PrivkeyBytes{}
	copy(b[:], (*secp256k1.PrivateKey)(prv).Serialize())
	return
}

// PrivkeyFromBytes converts a byte slice into a private key.
func PrivkeyFromBytes(b []byte) (prv *Privkey) {
	var p *secp256k1.PrivateKey
	p = secp256k1.PrivKeyFromBytes(b)
	prv = (*Privkey)(p)
	return
}

func (pb *PrivkeyBytes) Deserialize() (prv *Privkey) {
	return PrivkeyFromBytes(pb[:])
}

var zero PrivkeyBytes

// Zero zeroes out a private key in serial form. Note that sliced [:] form
// refers to the same bytes, so they are also zeroed.
func (pb *PrivkeyBytes) Zero() {
	copy(pb[:], zero[:])
}

// Serialize returns the compressed 33 byte form of the pubkey as for use with
// Schnorr signatures.
func (pub *Pubkey) Serialize() (p *PubkeyBytes) {
	p = &PubkeyBytes{}
	copy(p[:], (*secp256k1.PublicKey)(pub).SerializeCompressed())
	return
}

func (pb *PubkeyBytes) Deserialize() (pub *Pubkey, e error) {
	return PubkeyFromBytes(pb[:])
}

// ECDH computes an elliptic curve diffie hellman shared secret that can be
// decrypted by the holder of the private key matching the public key provided.
func (prv *Privkey) ECDH(pub *Pubkey) hasj.Hash {
	pr := (*secp256k1.PrivateKey)(prv)
	pu := (*secp256k1.PublicKey)(pub)
	b := sha256.Hash(secp256k1.GenerateSharedSecret(pr, pu))
	return b
}
