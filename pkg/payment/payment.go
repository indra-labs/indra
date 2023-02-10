package payment

import (
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"

	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
)

type Payment struct {
	nonce.ID
	Preimage sha256.Hash
	Amount   lnwire.MilliSatoshi
}
