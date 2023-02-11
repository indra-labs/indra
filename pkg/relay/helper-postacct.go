package relay

import (
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/onion"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/balance"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/confirm"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/crypt"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/exit"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/forward"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/getbalance"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/reverse"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type SendData struct {
	b         slice.Bytes
	sessions  traffic.Sessions
	billable  []nonce.ID
	ret, last nonce.ID
	port      uint16
	postAcct  []func()
}

// PostAcctOnion takes a slice of Skins and calculates their costs and
// the list of sessions inside them and attaches accounting operations to
// apply when the associated confirmation(s) or response hooks are executed.
func (eng *Engine) PostAcctOnion(o onion.Skins) (res SendData) {
	res.b = onion.Encode(o.Assemble())
	// do client accounting
	skip := false
	for i := range o {
		if skip {
			skip = false
			continue
		}
		switch on := o[i].(type) {
		case *crypt.Layer:
			s := eng.FindSessionByHeaderPub(on.ToHeaderPub)
			if s == nil {
				continue
			}
			res.sessions = append(res.sessions, s)
			// The last hop needs no accounting as it's us!
			if i == len(o)-1 {
				// The session used for the last hop is stored, however.
				res.ret = s.ID
				res.billable = append(res.billable, s.ID)
				break
			}
			switch on2 := o[i+1].(type) {
			case *forward.Layer:
				res.billable = append(res.billable, s.ID)
				res.postAcct = append(res.postAcct,
					func() {
						eng.DecSession(s.ID,
							s.RelayRate*
								lnwire.MilliSatoshi(len(res.b))/1024/1024, true,
							"forward")
					})
			case *reverse.Layer:
				res.billable = append(res.billable, s.ID)
			case *exit.Layer:
				for j := range s.Services {
					if s.Services[j].Port != on2.Port {
						continue
					}
					res.port = on2.Port
					res.postAcct = append(res.postAcct,
						func() {
							eng.DecSession(s.ID,
								s.Services[j].RelayRate*
									lnwire.MilliSatoshi(len(res.b)/2)/1024/1024,
								true, "exit")
						})
					break
				}
				res.billable = append(res.billable, s.ID)
				res.last = on2.ID
				skip = true
			case *getbalance.Layer:
				res.last = s.ID
				res.billable = append(res.billable, s.ID)
				skip = true
			}
		case *confirm.Layer:
			res.last = on.ID
		case *balance.Layer:
			res.last = on.ID
		}
	}
	return
}
