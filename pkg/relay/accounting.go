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
}

type PendingResponses struct {
	sync.Mutex
	responses     []*PendingResponse
	oldestPending *PendingResponse
	Timeout       time.Duration
}

func (p *PendingResponses) GetOldestPending() (pr *PendingResponse) {
	p.Lock()
	defer p.Unlock()
	return p.oldestPending
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
	log.T.F("deleting response %x", id)
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
	// Update the oldest pending response entry.
	if len(p.responses) > 0 {
		oldest := time.Now()
		for i := range p.responses {
			if p.responses[i].Time.Before(oldest) {
				oldest = p.responses[i].Time
				p.oldestPending = p.responses[i]
			}
		}
		// Add handler to trigger after timeout.
	} else {
		p.oldestPending = nil
	}
}
