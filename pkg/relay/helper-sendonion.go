package relay

import (
	"net/netip"
	
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/onion"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/balance"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/confirm"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/crypt"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/directbalance"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/exit"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/forward"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/getbalance"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/reverse"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (eng *Engine) SendOnion(ap *netip.AddrPort, o onion.Skins,
	responseHook func(id nonce.ID, b slice.Bytes)) {
	
	b := onion.Encode(o.Assemble())
	var billable, accounted []nonce.ID
	var ret nonce.ID
	var last nonce.ID
	var port uint16
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
			// The last hop needs no accounting as it's us!
			if i == len(o)-1 {
				// The session used for the last hop is stored, however.
				ret = s.ID
				break
			}
			if s == nil {
				continue
			}
			switch on2 := o[i+1].(type) {
			case *forward.Layer:
				eng.DecSession(s.ID,
					s.RelayRate*lnwire.MilliSatoshi(len(b))/1024/1024, true,
					"forward")
				accounted = append(accounted, s.ID)
			case *reverse.Layer:
				billable = append(billable, s.ID)
			case *exit.Layer:
				for j := range s.Services {
					if s.Services[j].Port != on2.Port {
						continue
					}
					port = on2.Port
					eng.DecSession(s.ID,
						s.Services[j].RelayRate*lnwire.
							MilliSatoshi(len(b)/2)/1024/1024, true, "exit")
					accounted = append(accounted, s.ID)
					break
				}
				billable = append(billable, s.ID)
				last = on2.ID
				skip = true
			case *getbalance.Layer:
				eng.DecSession(s.ID,
					s.RelayRate*lnwire.MilliSatoshi(len(b)/2)/1024/1024, true,
					"getbalance")
				last = s.ID
				billable = append(billable, s.ID)
				skip = true
			}
		case *directbalance.Layer:
			// the immediate previous layer session needs to be accounted.
			switch on3 := o[i-1].(type) {
			case *crypt.Layer:
				s := eng.FindSessionByHeaderPub(on3.ToHeaderPub)
				if s == nil {
					return
				}
				last = on.ID
			}
		case *confirm.Layer:
			last = on.ID
		case *balance.Layer:
			last = on.ID
		}
	}
	if responseHook == nil {
		responseHook = func(_ nonce.ID, _ slice.Bytes) {
			log.T.Ln("nil response hook")
		}
	}
	eng.PendingResponses.Add(last, len(b), billable, accounted, ret, port, responseHook)
	log.T.Ln("sending out onion")
	eng.Send(ap, b)
	
}
