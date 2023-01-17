package payment

import (
	"github.com/indra-labs/indra/pkg/lnd/lnwire"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/sha256"
)

type Payment struct {
	nonce.ID
	Preimage sha256.Hash
	Amount   lnwire.MilliSatoshi
}
