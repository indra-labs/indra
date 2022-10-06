package ciph

import (
	"github.com/Indra-Labs/indra/pkg/schnorr"
)

type Message struct {
	PubKey    schnorr.PubkeyBytes
	Nonce     Nonce
	Message   []byte
	Signature schnorr.SignatureBytes
}
