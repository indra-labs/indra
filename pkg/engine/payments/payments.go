// Package payments provides an abstraction above the implementation for handling Lightning Network payments and storing pending payments awaiting session keys.
package payments

import (
	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
	"git.indra-labs.org/dev/ind/pkg/crypto/sha256"
	"github.com/lightningnetwork/lnd/lnwire"
)

// Add a new payment to the PendingPayments and return the mutated slice.
func (p PendingPayments) Add(np *Payment) (pp PendingPayments) {
	return append(p, np)
}

type (
	// PayChan is a channel for sending and receiving payments between the LN
	// interface and the Engine.
	PayChan chan *Payment

	// Payment is the essential details of a payment and the method to wait for full
	// confirmation.
	Payment struct {

		// ID is the unique identifier that will be used internally also for the session
		// when the session onion arrives.
		ID nonce.ID

		// Preimage is the hash of the pair of secret keys that proves a session key set
		// should be tied to the provided secret keys for decrypting onions for this
		// session.
		Preimage sha256.Hash

		// Amount in millisatoshi that have been received over LN.
		Amount lnwire.MilliSatoshi

		// ConfirmChan signals from the LN server to the Engine that the payment is
		// complete, and if available the session can now be spent from.
		ConfirmChan chan bool
	}

	// PendingPayments is a slice of Payments used to store the queue of payments
	// awaiting sessions.
	PendingPayments []*Payment
)

// Delete a pending payment and return the mutated slice with the item matching
// the preimage removed.
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

// Find the pending payment with the matching nonce.ID (internal reference).
func (p PendingPayments) Find(id nonce.ID) (pp *Payment) {
	for i := range p {
		if p[i].ID == id {
			return p[i]
		}
	}
	return
}

// FindPreimage searches for a match for the preimage hash, as would be needed
// after receiving the preimage in the payment the session's keys derive the
// preimage for this search.
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
