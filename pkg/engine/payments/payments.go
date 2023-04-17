package payments

import (
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
)

type Payment struct {
	ID          nonce.ID
	Preimage    sha256.Hash
	Amount      lnwire.MilliSatoshi
	ConfirmChan chan bool
}

type Chan chan *Payment

// Send a payment on the Chan.
func (pc Chan) Send(amount lnwire.MilliSatoshi,
	id nonce.ID, preimage sha256.Hash, ) (confirmChan chan bool) {
	
	confirmChan = make(chan bool)
	pc <- &Payment{
		ID:          id,
		Preimage:    preimage,
		Amount:      amount,
		ConfirmChan: confirmChan,
	}
	return
}

// Receive waits on receiving a Payment on a Chan.
func (pc Chan) Receive() <-chan *Payment { return pc }

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
