package schnorr

import (
	"fmt"

	"github.com/Indra-Labs/indra/pkg/schnorr/schnorr"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

const (
	PrivkeyLen     = secp256k1.PrivKeyBytesLen
	PubkeyLen      = schnorr.PubKeyBytesLen
	SigLen         = schnorr.SignatureSize
	FingerprintLen = 8
	HashLen        = 32
)

// Privkey is a private key
type Privkey secp256k1.PrivateKey
type PrivkeyBytes [PrivkeyLen]byte

// Pubkey is a public key
type Pubkey secp256k1.PublicKey
type PubkeyBytes [PubkeyLen]byte

// Signature is a schnorr signature over secp256k1 curve
type Signature schnorr.Signature
type SignatureBytes [SigLen]byte

// Fingerprint is a truncated SHA256D hash of the pubkey, indicating the relevant
// key when the full Pubkey will be available, and for easier human recognition.
type Fingerprint [FingerprintLen]byte

// Hash is just a byte type with a nice length validation function
type Hash []byte

func (h Hash) Valid() error {
	if len(h) == HashLen {
		return nil
	}
	return fmt.Errorf("invalid hash length of %d bytes, must be %d",
		len(h), HashLen)
}

func (h Hash) Zero() {
	for i := range h {
		h[i] = 0
	}
}

// Fingerprint generates a fingerprint from a Pubkey
func (pub Pubkey) Fingerprint() (fp Fingerprint) {
	h := sha256.Double(pub.Serialize()[:])
	copy(fp[:], h[:FingerprintLen])
	return
}

func (pub PubkeyBytes) Fingerprint() (fp *Fingerprint) {
	fp = &Fingerprint{}
	h := sha256.Double(pub[:])
	copy(fp[:], h[:FingerprintLen])
	return

}

// GeneratePrivkey generates a private key
func GeneratePrivkey() (prv *Privkey, e error) {
	var p *secp256k1.PrivateKey
	if p, e = secp256k1.GeneratePrivateKey(); log.I.Chk(e) {
		return
	}
	prv = (*Privkey)(p)
	return
}

// Zero zeroes out a private key to prevent key scraping from memory
func (prv *Privkey) Zero() {
	(*secp256k1.PrivateKey)(prv).Zero()
}

// Pubkey generates a public key from the Privkey
func (prv *Privkey) Pubkey() *Pubkey {
	return (*Pubkey)((*secp256k1.PrivateKey)(prv).PubKey())
}

// Serialize returns the PrivkeyBytes serialized form
func (prv *Privkey) Serialize() (b *PrivkeyBytes) {
	b = &PrivkeyBytes{}
	copy(b[:], (*secp256k1.PrivateKey)(prv).Serialize())
	return
}

// PrivkeyFromBytes converts a byte slice into a private key
func PrivkeyFromBytes(b []byte) (prv *Privkey) {
	var p *secp256k1.PrivateKey
	p = secp256k1.PrivKeyFromBytes(b)
	prv = (*Privkey)(p)
	return
}

func (prv *PrivkeyBytes) Deserialize() (priv *Privkey) {
	return PrivkeyFromBytes(prv[:])
}

var zero PrivkeyBytes

// Zero zeroes out a private key in serial form. Note that sliced [:] form
// refers to the same bytes so they are also zeroed (todo: check this is true)
func (p *PrivkeyBytes) Zero() {
	copy(p[:], zero[:])
}

// Serialize returns the compressed 33 byte form of the pubkey as for use with
// Schnorr signatures.
func (pub *Pubkey) Serialize() (p *PubkeyBytes) {
	p = &PubkeyBytes{}
	copy(p[:], (*secp256k1.PublicKey)(pub).SerializeCompressed())
	return
}

func (pk *PubkeyBytes) Deserialize() (pub *Pubkey, e error) {
	return PubkeyFromBytes(pk[:])
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

// ECDH computes an elliptic curve diffie hellman shared secret that can be
// decrypted by the holder of the private key matching the public key provided.
func (prv *Privkey) ECDH(pub *Pubkey) Hash {
	pr := (*secp256k1.PrivateKey)(prv)
	pu := (*secp256k1.PublicKey)(pub)
	b := sha256.Double(secp256k1.GenerateSharedSecret(pr, pu))
	return b
}

func (prv *Privkey) Sign(hash Hash) (sig *Signature, e error) {
	if log.E.Chk(hash.Valid()) {
		return
	}
	var s *schnorr.Signature
	if s, e = schnorr.Sign((*secp256k1.PrivateKey)(prv), hash); log.E.Chk(e) {
		return
	}
	sig = (*Signature)(s)
	return
}

func ParseSignature(s []byte) (sig *Signature, e error) {
	var ss *schnorr.Signature
	if ss, e = schnorr.ParseSignature(s); log.E.Chk(e) {
		return
	}
	sig = (*Signature)(ss)
	return
}

func (sig *Signature) Verify(hash Hash, pub *Pubkey) bool {
	if log.E.Chk(hash.Valid()) {
		return false
	}
	return (*schnorr.Signature)(sig).
		Verify(hash, (*secp256k1.PublicKey)(pub))
}

func (sig *Signature) Serialize() (s *SignatureBytes) {
	s = &SignatureBytes{}
	copy(s[:], (*schnorr.Signature)(sig).Serialize())
	return
}

func (sb SignatureBytes) Deserialize() (sig *Signature, e error) {
	if sig, e = ParseSignature(sb[:]); log.E.Chk(e) {
	}
	return
}
