// Package cert is an interface for messages bearing signatures of network participants.
package cert

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/crypto"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

// Act is an interface for the signed messages stored in the PeerStore of the
// libp2p host inside an indra engine.
//
// Note the abstract name and the open ended possibility of composing this into
// other types of contract documents. In the language of Law an Act is the
// prototype of a declaration or claim, a land title is an example of a type of
// Act.
//
// In Indra, this is a spam-resistant message type that invites validation and
// indirectly spammy abuse of this message type in the gossip network will cause
// banning and the eviction of the record.
//
// Hidden services introduction advertisement is an example of the hidden
// service attesting to the provision of the referral messages found in the
// package: pkg/codec/onion/hidden/services - These are expected to become far
// more numerous than peer advertisments as they effectively designate a
// listening server. These can be spam-controlled by having peers poke at the
// service and dropping non-working intros and thus potentially leading to the
// hidden service intro being evicted from the collective peerstore.
//
// todo: hidden service sessions...
type Act interface {
	codec.Codec
	Sign(key *crypto.Prv) (e error)
	Validate() bool
}
