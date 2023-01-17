package identity

import (
	"net/netip"

	"github.com/indra-labs/indra/pkg/crypto/key/prv"
	"github.com/indra-labs/indra/pkg/crypto/key/pub"
	"github.com/indra-labs/indra/pkg/types"
)

type Peer struct {
	AddrPort      *netip.AddrPort
	IdentityPub   *pub.Key
	IdentityBytes pub.Bytes
	IdentityPrv   *prv.Key
	types.Transport
}
