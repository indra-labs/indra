package relay

import (
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/relay/messages/session"
)

type Payment struct {
	nonce.ID
	Preimage    sha256.Hash
	Amount      lnwire.MilliSatoshi
	ConfirmChan chan bool
}

type PaymentChan chan *Payment

// Send a payment on the PaymentChan.
func (pc PaymentChan) Send(amount lnwire.MilliSatoshi,
	s *session.Layer) (confirmChan chan bool) {
	
	confirmChan = make(chan bool)
	pc <- &Payment{
		ID:          s.ID,
		Preimage:    s.PreimageHash(),
		Amount:      amount,
		ConfirmChan: confirmChan,
	}
	return
}

// Receive waits on receiving a Payment on a PaymentChan.
func (pc PaymentChan) Receive() <-chan *Payment { return pc }

type PendingPayments []*Payment

func (p PendingPayments) Add(np *Payment) (pp PendingPayments) {
	return append(p, np)
}

func (p PendingPayments) Delete(preimage sha256.Hash) (pp PendingPayments) {
	pp = p
	for i := range p {
		if p[i].Preimage == preimage {
			if i == len(p)-1 {
				pp = p[:i]
			} else {
				pp = append(p[:i], p[i+1:]...)
			}
		}
	}
	return
}

func (p PendingPayments) Find(id nonce.ID) (pp *Payment) {
	for i := range p {
		if p[i].ID == id {
			return p[i]
		}
	}
	return
}

func (p PendingPayments) FindPreimage(pi sha256.Hash) (pp *Payment) {
	for i := range p {
		if p[i].Preimage == pi {
			return p[i]
		}
	}
	return
}

// PendingPayment accessors. For the same reason as the sessions, pending
// payments need to be accessed only with the node's mutex locked.

func (sm *SessionManager) AddPendingPayment(np *Payment) {
	sm.Lock()
	defer sm.Unlock()
	log.D.F("%s adding pending payment %s for %v",
		sm.nodes[0].AddrPort.String(), np.ID,
		np.Amount)
	sm.PendingPayments = sm.PendingPayments.Add(np)
}
func (sm *SessionManager) DeletePendingPayment(preimage sha256.Hash) {
	sm.Lock()
	defer sm.Unlock()
	sm.PendingPayments = sm.PendingPayments.Delete(preimage)
}
func (sm *SessionManager) FindPendingPayment(id nonce.ID) (pp *Payment) {
	sm.Lock()
	defer sm.Unlock()
	return sm.PendingPayments.Find(id)
}
func (sm *SessionManager) FindPendingPreimage(pi sha256.Hash) (pp *Payment) {
	log.T.F("searching preimage %x", pi)
	sm.Lock()
	defer sm.Unlock()
	return sm.PendingPayments.FindPreimage(pi)
}
