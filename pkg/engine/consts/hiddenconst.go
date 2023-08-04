// Package consts is a series of constants common to several different onion message types.
package consts

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/util/splice"
)

// ReverseCryptLen is
//
// Deprecated: this is now a variable length structure, Reverse is being
// obsoleted in favour of Offset.
const ReverseCryptLen = ReverseLen + CryptLen

// RoutingHeaderLen is
//
// Deprecated: this is now a variable length structure.
const RoutingHeaderLen = 3 * ReverseCryptLen

const CryptLen = magic.Len +
	nonce.IVLen +
	crypto.CloakLen +
	crypto.PubKeyLen

// ReverseLen is
//
// Deprecated: Reverse is being obsoleted in favour of Offset.
const ReverseLen = magic.Len + 1 + splice.AddrLen
