package schnorr

import (
	"crypto/sha256"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/schnorr"
	"lukechampine.com/blake3"
)

const (
	PrivkeyLen     = secp256k1.PrivKeyBytesLen
	PubkeyLen      = schnorr.PubKeyBytesLen
	SigLen         = schnorr.SignatureSize
	FingerprintLen = 8
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

// Fingerprint is a truncated SHA256 hash of the pubkey, indicating the relevant
// key when the full Pubkey will be available, and for easier human recognition.
type Fingerprint [FingerprintLen]byte

// SHA256D runs a standard double SHA256 hash and does all the slicing for you.
func SHA256D(b []byte) []byte {
	h := sha256.Sum256(b)
	h = sha256.Sum256(h[:])
	return h[:]
}

// Fingerprint generates a fingerprint from a Pubkey
func (pub Pubkey) Fingerprint() (fp Fingerprint) {
	h := SHA256D(pub.Serialize()[:])
	copy(fp[:], h[:FingerprintLen])
	return
}

// GeneratePrivkey generates a private key
func GeneratePrivkey() (prv *Privkey, err error) {
	var p *secp256k1.PrivateKey
	if p, err = secp256k1.GeneratePrivateKey(); log.I.Chk(err) {
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

// PubkeyFromBytes converts a byte slice into a public key, if it is valid.
func PubkeyFromBytes(b []byte) (pub *Pubkey, err error) {
	var p *secp256k1.PublicKey
	p, err = secp256k1.ParsePubKey(b)
	pub = (*Pubkey)(p)
	return
}

func (prv *Privkey) ECDH(pub *Pubkey) []byte {
	pr := (*secp256k1.PrivateKey)(prv)
	pu := (*secp256k1.PublicKey)(pub)
	b := SHA256D(secp256k1.GenerateSharedSecret(pr, pu))
	return b
}

func (prv *Privkey) Sign(message []byte) (sig *Signature, err error) {
	hash := blake3.Sum256(message)
	var s *schnorr.Signature
	s, err = schnorr.Sign((*secp256k1.PrivateKey)(prv), hash[:])
	sig = (*Signature)(s)
	return
}

func ParseSignature(s []byte) (sig *Signature, err error) {
	var ss *schnorr.Signature
	ss, err = schnorr.ParseSignature(s)
	sig = (*Signature)(ss)
	return
}

func (sig *Signature) Verify(message []byte, pub *Pubkey) bool {
	hash := blake3.Sum256(message)
	return (*schnorr.Signature)(sig).Verify(hash[:],
		(*secp256k1.PublicKey)(pub))
}

func (sig *Signature) Serialize() (s *SignatureBytes) {
	s = &SignatureBytes{}
	copy(s[:], (*schnorr.Signature)(sig).Serialize())
	return
}
