package payment

import (
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/lnd/lnd/lnwire"
)

type Payment struct {
	nonce.ID
	Preimage sha256.Hash
	Amount   lnwire.MilliSatoshi
}
