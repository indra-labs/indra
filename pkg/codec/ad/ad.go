// Package ad is an abstract message type that composes the common elements of all ads - nonce ID, public key (identity), expiry and signature.
//
// The concrete ad types are in subfolders of this package.
package ad

import (
	"git.indra-labs.org/dev/ind/pkg/crypto"
	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
	"git.indra-labs.org/dev/ind/pkg/engine/magic"
	"git.indra-labs.org/dev/ind/pkg/util/slice"
	"time"
)

const (
	Len = magic.Len +
		nonce.IDLen +
		crypto.PubKeyLen +
		slice.Uint64Len +
		crypto.SigLen
)

// Ad is an abstract message that isn't actually used, but rather is composed
// into the rest being the common fields to all ads.
type Ad struct {

	// ID is a random number generated for each new add that practically guarantees
	// no repeated hash for a message to be signed with the same peer identity key.
	ID nonce.ID

	// Key is the public identity key of the peer.
	Key *crypto.Pub

	// Expiry is the time after which this ad is considered no longer current.
	Expiry time.Time

	// Sig is the signature, which must match the Key above.
	Sig crypto.SigBytes
}
