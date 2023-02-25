package relay

import (
	"net/netip"
	"time"
	
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/onion/balance"
	"git-indra.lan/indra-labs/indra/pkg/onion/confirm"
	"git-indra.lan/indra-labs/indra/pkg/onion/crypt"
	"git-indra.lan/indra-labs/indra/pkg/onion/exit"
	"git-indra.lan/indra-labs/indra/pkg/onion/forward"
	"git-indra.lan/indra-labs/indra/pkg/onion/getbalance"
	"git-indra.lan/indra-labs/indra/pkg/onion/reverse"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (eng *Engine) SendOnion(ap *netip.AddrPort, o Skins,
	responseHook func(id nonce.ID, b slice.Bytes), timeout time.Duration) {
	
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	b := Encode(o.Assemble())
	var billable []nonce.ID
	var ret nonce.ID
	var last nonce.ID
	var port uint16
	var postAcct []func()
	var sessions Sessions
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
			sessions = append(sessions, s)
			// The last hop needs no accounting as it's us!
			if i == len(o)-1 {
				// The session used for the last hop is stored, however.
				ret = s.ID
				billable = append(billable, s.ID)
				break
			}
			switch on2 := o[i+1].(type) {
			case *forward.Layer:
				billable = append(billable, s.ID)
				postAcct = append(postAcct,
					func() {
						eng.DecSession(s.ID,
							s.RelayRate*
								lnwire.MilliSatoshi(len(b))/1024/1024, true,
							"forward")
					})
			case *reverse.Layer:
				billable = append(billable, s.ID)
			case *exit.Layer:
				for j := range s.Services {
					if s.Services[j].Port != on2.Port {
						continue
					}
					port = on2.Port
					postAcct = append(postAcct,
						func() {
							eng.DecSession(s.ID,
								s.Services[j].RelayRate*
									lnwire.MilliSatoshi(len(b)/2)/1024/1024,
								true, "exit")
						})
					break
				}
				billable = append(billable, s.ID)
				last = on2.ID
				skip = true
			case *getbalance.Layer:
				last = s.ID
				billable = append(billable, s.ID)
				skip = true
			}
		case *confirm.Layer:
			last = on.ID
		case *balance.Layer:
			last = on.ID
		}
	}
	if responseHook == nil {
		responseHook = func(_ nonce.ID, _ slice.Bytes) {
			log.D.Ln("nil response hook")
		}
	}
	eng.PendingResponses.Add(last, len(b), sessions, billable, ret, port,
		responseHook, postAcct)
	log.T.Ln("sending out onion")
	eng.Send(ap, b)
}
