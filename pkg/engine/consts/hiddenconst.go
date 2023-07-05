// Package consts is a series of constants common to several different onion message types.
package consts

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/util/splice"
)

const ReverseCryptLen = ReverseLen + CryptLen

const RoutingHeaderLen = 3 * ReverseCryptLen

const CryptLen = magic.Len +
	nonce.IVLen +
	crypto.CloakLen +
	crypto.PubKeyLen

const ReverseLen = magic.Len + 1 + splice.AddrLen
