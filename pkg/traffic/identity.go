package traffic

import (
	"net/netip"

	"github.com/indra-labs/indra/pkg/crypto/key/prv"
	"github.com/indra-labs/indra/pkg/crypto/key/pub"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/ring"
	"github.com/indra-labs/indra/pkg/service"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/lnd/lnd/lnwire"
)

type Peer struct {
	nonce.ID      // matches node ID
	AddrPort      *netip.AddrPort
	IdentityPub   *pub.Key
	IdentityBytes pub.Bytes
	IdentityPrv   *prv.Key
	RelayRate     lnwire.MilliSatoshi // Base relay price/Mb.
	Services      service.Services    // Services offered by this peer.
	Load          *ring.BufferLoad    // Relay load.
	Latency       *ring.BufferLatency // Latency to peer.
	Failure       *ring.BufferFailure // Times of tx failure.
	types.Transport
	*Payments
}
