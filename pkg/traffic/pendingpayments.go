package traffic

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/payment"
)

type PendingPayments []*payment.Payment

func (p PendingPayments) Add(np *payment.Payment) (pp PendingPayments) {
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

func (p PendingPayments) Find(id nonce.ID) (pp *payment.Payment) {
	for i := range p {
		if p[i].ID == id {
			return p[i]
		}
	}
	return
}

func (p PendingPayments) FindPreimage(pi sha256.Hash) (pp *payment.Payment) {
	for i := range p {
		if p[i].Preimage == pi {
			return p[i]
		}
	}
	return
}

// PendingPayment accessors. For the same reason as the sessions, pending
// payments need to be accessed only with the node's mutex locked.

func (sm *SessionManager) AddPendingPayment(
	np *payment.Payment) {

	sm.Lock()
	defer sm.Unlock()
	sm.pendingPayments = sm.pendingPayments.Add(np)
}
func (sm *SessionManager) DeletePendingPayment(
	preimage sha256.Hash) {

	sm.Lock()
	defer sm.Unlock()
	sm.pendingPayments = sm.pendingPayments.Delete(preimage)
}
func (sm *SessionManager) FindPendingPayment(
	id nonce.ID) (pp *payment.Payment) {

	sm.Lock()
	defer sm.Unlock()
	return sm.pendingPayments.Find(id)
}
func (sm *SessionManager) FindPendingPreimage(
	pi sha256.Hash) (pp *payment.Payment) {

	log.T.F("searching preimage %x", pi)
	sm.Lock()
	defer sm.Unlock()
	return sm.pendingPayments.FindPreimage(pi)
}
