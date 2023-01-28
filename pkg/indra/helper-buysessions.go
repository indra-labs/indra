package indra

import (
	"github.com/indra-labs/lnd/lnd/lnwire"

	"github.com/indra-labs/indra/pkg/node"
	"github.com/indra-labs/indra/pkg/onion/layers/session"
	"github.com/indra-labs/indra/pkg/payment"
)

func (en *Engine) BuySessions(
	s ...*SessionBuy) (sess []*session.Layer,
	pmt []*payment.Payment) {

	for i := range s {
		// Create a new payment and drop on the payment channel.
		sess = append(sess, session.New(s[i].Hop))
		pmt = append(pmt, sess[i].ToPayment(s[i].Amount))
		s[i].PaymentChan <- pmt[i]
		log.T.Ln("sent out payment", i)
	}
	return
}

func BuySession(n *node.Node, amt lnwire.MilliSatoshi, hop byte) (o *SessionBuy) {
	return &SessionBuy{Hop: hop, Amount: amt, Node: n}
}

type SessionBuy struct {
	Hop    byte
	Amount lnwire.MilliSatoshi
	*node.Node
}
