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

func (pm *Payments) AddPendingPayment(
	np *payment.Payment) {

	pm.Lock()
	defer pm.Unlock()
	pm.pendingPayments = pm.pendingPayments.Add(np)
}
func (pm *Payments) DeletePendingPayment(
	preimage sha256.Hash) {

	pm.Lock()
	defer pm.Unlock()
	pm.pendingPayments = pm.pendingPayments.Delete(preimage)
}
func (pm *Payments) FindPendingPayment(
	id nonce.ID) (pp *payment.Payment) {

	pm.Lock()
	defer pm.Unlock()
	return pm.pendingPayments.Find(id)
}
func (pm *Payments) FindPendingPreimage(
	pi sha256.Hash) (pp *payment.Payment) {

	log.T.F("searching preimage %x", pi)
	pm.Lock()
	defer pm.Unlock()
	return pm.pendingPayments.FindPreimage(pi)
}
