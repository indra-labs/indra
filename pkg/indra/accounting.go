package indra

import (
	"sync"
	"time"

	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/util/slice"
)

type Hook func(cf sha256.Hash)

type PendingResponse struct {
	sha256.Hash
	Port                uint16
	Billable, Accounted []nonce.ID
	Return              nonce.ID
	Callback            func(b slice.Bytes)
	time.Time
}

type PendingResponses struct {
	sync.Mutex
	responses     []*PendingResponse
	oldestPending *PendingResponse
}

func (p *PendingResponses) GetOldestPending() (pr *PendingResponse) {
	p.Lock()
	defer p.Unlock()
	return p.oldestPending
}

func (p *PendingResponses) Add(h sha256.Hash, billable, accounted []nonce.ID,
	ret nonce.ID, port uint16, callback func(b slice.Bytes)) {

	p.Lock()
	defer p.Unlock()
	log.T.F("adding response hook %x", h)
	p.responses = append(p.responses, &PendingResponse{
		Hash:      h,
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

func (p *PendingResponses) Find(h sha256.Hash) (pr *PendingResponse) {
	p.Lock()
	defer p.Unlock()
	for i := range p.responses {
		if p.responses[i].Hash == h {

			return p.responses[i]
		}
	}
	return
}

func (p *PendingResponses) Delete(h sha256.Hash) {
	p.Lock()
	defer p.Unlock()
	for i := range p.responses {
		if p.responses[i].Hash == h {
			if i < 1 {
				p.responses = p.responses[1:]
			} else {
				p.responses = append(p.responses[:i],
					p.responses[i+1:]...)
			}
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
	} else {
		p.oldestPending = nil
	}
}
