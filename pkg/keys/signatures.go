package keys

import (
	"github.com/Indra-Labs/indra/pkg/hasj"
	"github.com/Indra-Labs/indra/pkg/keys/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Signature is a schnorr signature over secp256k1 curve
type Signature schnorr.Signature
type SignatureBytes [SigLen]byte

func (prv *Privkey) Sign(hash hasj.Hash) (sig *Signature, e error) {
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

func (sig *Signature) Verify(hash hasj.Hash, pub *Pubkey) bool {
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

const SigLen = schnorr.SignatureSize
