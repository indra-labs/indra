package payments

import (
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/lightningnetwork/lnd/lnwire"
)

func (p PendingPayments) Add(np *Payment) (pp PendingPayments) {
	return append(p, np)
}

type (
	PayChan chan *Payment
	Payment struct {
		ID          nonce.ID
		Preimage    sha256.Hash
		Amount      lnwire.MilliSatoshi
		ConfirmChan chan bool
	}
	PendingPayments []*Payment
)

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

// Receive waits on receiving a Payment on a PayChan.
func (pc PayChan) Receive() <-chan *Payment { return pc }

// Send a payment on the PayChan.
func (pc PayChan) Send(amount lnwire.MilliSatoshi,
	id nonce.ID, preimage sha256.Hash) (confirmChan chan bool) {
	confirmChan = make(chan bool)
	pc <- &Payment{
		ID:          id,
		Preimage:    preimage,
		Amount:      amount,
		ConfirmChan: confirmChan,
	}
	return
}
