// Package cert is an interface for messages bearing signatures of network participants.
package cert

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

// Act is an interface for the signed messages stored in the PeerStore of the
// libp2p host inside an indra engine.
//
// Note the abstract name and the open ended possibility of composing this into
// other types of contract documents. In the language of Law an Act is the
// prototype of a declaration or claim, a land title is an example of a type of
// Act.
type Act interface {
	codec.Codec
	Sign(key *crypto.Prv) (e error)
	Validate() bool
	PubKey() (pubKey *crypto.Pub)
	GetID() (id peer.ID, e error)
	Expired() (is bool)
}
