// Package consts is a series of constants common to several different onion message types.
package consts

import (
	"git.indra-labs.org/dev/ind/pkg/crypto"
	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
	"git.indra-labs.org/dev/ind/pkg/engine/magic"
	"git.indra-labs.org/dev/ind/pkg/util/splice"
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
