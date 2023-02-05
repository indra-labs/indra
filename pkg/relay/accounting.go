package relay

import (
	"sync"
	"time"

	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type Callback func(id nonce.ID, b slice.Bytes)

type PendingResponse struct {
	nonce.ID
	SentSize            int
	Port                uint16
	Billable, Accounted []nonce.ID
	Return              nonce.ID
	Callback
	time.Time
	Timeout time.Duration
}

type PendingResponses struct {
	sync.Mutex
	responses []*PendingResponse
}

func (p *PendingResponses) GetOldestPending() (pr *PendingResponse) {
	p.Lock()
	defer p.Unlock()
	if len(p.responses) > 0 {
		// Pending responses are added in chronological order to the end so the
		// first one in the slice is the oldest.
		return p.responses[0]
	}
	return
}

func (p *PendingResponses) Add(id nonce.ID, sentSize int, billable,
	accounted []nonce.ID, ret nonce.ID, port uint16,
	callback func(id nonce.ID, b slice.Bytes)) {

	p.Lock()
	defer p.Unlock()
	log.T.F("adding response hook %x", id)
	p.responses = append(p.responses, &PendingResponse{
		ID:        id,
		SentSize:  sentSize,
		Time:      time.Now(),
		Billable:  billable,
		Accounted: accounted,
		Return:    ret,
		Port:      port,
		Callback:  callback,
	})
}

func (p *PendingResponses) FindOlder(t time.Time) (r []*PendingResponse) {
	p.Lock()
	defer p.Unlock()
	for i := range p.responses {
		if p.responses[i].Time.Before(t) {
			r = append(r, p.responses[i])
		}
	}
	return
}

func (p *PendingResponses) Find(id nonce.ID) (pr *PendingResponse) {
	p.Lock()
	defer p.Unlock()
	for i := range p.responses {
		if p.responses[i].ID == id {
			return p.responses[i]
		}
	}
	return
}

func (p *PendingResponses) Delete(id nonce.ID, b slice.Bytes) {
	p.Lock()
	defer p.Unlock()
	log.T.F("deleting response %s", id)
	for i := range p.responses {
		if p.responses[i].ID == id {
			p.responses[i].Callback(id, b)
			if i < 1 {
				p.responses = p.responses[1:]
			} else {
				p.responses = append(p.responses[:i],
					p.responses[i+1:]...)
			}
			break
		}
	}
}
