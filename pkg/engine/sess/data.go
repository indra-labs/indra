package sess

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/util/slice"
)

type Data struct {
	B        slice.Bytes
	Sessions sessions.Sessions
	Billable []crypto.PubBytes
	Ret      crypto.PubBytes
	ID       nonce.ID
	Port     uint16
	PostAcct []func()
}
