package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type Data struct {
	B        slice.Bytes
	Sessions sessions.Sessions
	Billable []nonce.ID
	Ret, ID  nonce.ID
	Port     uint16
	PostAcct []func()
}
