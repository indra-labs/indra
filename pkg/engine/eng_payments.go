package engine

import (
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
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
	s *Session) (confirmChan chan bool) {
	
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
