// Package ad is an interface for peer information advertisements.
package ad

import (
	"github.com/indra-labs/indra/pkg/codec"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/splice"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

// Ad is an interface for the signed messages stored in the PeerStore of the
// libp2p host inside an indra engine.
type Ad interface {
	codec.Codec
	Splice(s *splice.Splice)
	Validate() bool
}
