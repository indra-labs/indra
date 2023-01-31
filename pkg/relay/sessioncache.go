package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
)

type SessionCache map[nonce.ID]SessionCacheEntry

type SessionCacheEntry struct {
	*traffic.Node
	Hops [5]*traffic.Session
}

func (eng *Engine) UpdateSessionCache() {
	for i := range eng.Nodes {
		_ = i
	}
}
