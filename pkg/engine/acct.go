package engine

import (
	"github.com/indra-labs/indra/pkg/codec/onion/crypt"
	"github.com/indra-labs/indra/pkg/codec/ont"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
)

// PostAcctOnion takes a slice of Skins and calculates their costs and
// the list of sessions inside them and attaches accounting operations to
// apply when the associated confirmation(s) or response hooks are executed.
func PostAcctOnion(sm *sess.Manager, o Skins) (res *sess.Data) {
	res = &sess.Data{}
	assembled := ont.Assemble(o)
	sp := ont.Encode(assembled)
	res.B = sp.GetAll()
	// do client accounting
	skip := false
	var last bool
	for i := range o {
		if skip {
			skip = false
			continue
		}
		switch on := o[i].(type) {
		case *crypt.Crypt:
			if i == len(o)-1 {
				last = true
			}
			var s *sessions.Data
			skip, s = on.Account(res, sm, nil, last)
			if last {
				break
			}
			o[i+1].Account(res, sm, s, last)
		}
	}
	return
}
