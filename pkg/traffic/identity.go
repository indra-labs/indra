package traffic

import (
	"net/netip"

	"git-indra.lan/indra-labs/lnd/lnd/lnwire"

	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/ring"
	"git-indra.lan/indra-labs/indra/pkg/service"
	"git-indra.lan/indra-labs/indra/pkg/types"
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
