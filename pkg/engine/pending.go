package engine

import (
	"sync"
	"time"
	
	"github.com/cybriq/qu"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type Callback func(id nonce.ID, b slice.Bytes)

type PendingResponse struct {
	nonce.ID
	SentSize int
	Port     uint16
	Billable []nonce.ID
	Return   nonce.ID
	PostAcct []func()
	Sessions
	Callback
	time.Time
	Success qu.C
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

type ResponseParams struct {
	ID       nonce.ID
	SentSize int
	S        Sessions
	Billable []nonce.ID
	Ret      nonce.ID
	Port     uint16
	Callback Callback
	PostAcct []func()
}

func (p *PendingResponses) Add(pr ResponseParams) {
	p.Lock()
	defer p.Unlock()
	log.T.F("adding response hook %s", pr.ID)
	r := &PendingResponse{
		ID:       pr.ID,
		SentSize: pr.SentSize,
		Time:     time.Now(),
		Billable: pr.Billable,
		Return:   pr.Ret,
		Port:     pr.Port,
		PostAcct: pr.PostAcct,
		Callback: pr.Callback,
		Success:  qu.T(),
	}
	p.responses = append(p.responses, r)
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

// ProcessAndDelete runs the callback and post accounting function list and
// deletes the pending response.
func (p *PendingResponses) ProcessAndDelete(id nonce.ID, b slice.Bytes) {
	p.Lock()
	defer p.Unlock()
	log.T.F("deleting response %s", id)
	for i := range p.responses {
		if p.responses[i].ID == id {
			// Stop the timeout handler.
			p.responses[i].Success.Q()
			for _, fn := range p.responses[i].PostAcct {
				fn()
			}
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
