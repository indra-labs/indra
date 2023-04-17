package responses

import (
	"sync"
	"time"
	
	"github.com/cybriq/qu"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

type Callback func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error)

type Response struct {
	ID       nonce.ID
	SentSize int
	Port     uint16
	Billable []nonce.ID
	Return   nonce.ID
	PostAcct []func()
	sessions.Sessions
	Callback Callback
	time.Time
	Success qu.C
}

type Pending struct {
	sync.Mutex
	responses []*Response
}

func (p *Pending) GetOldestPending() (pr *Response) {
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
	S        sessions.Sessions
	Billable []nonce.ID
	Ret      nonce.ID
	Port     uint16
	Callback Callback
	PostAcct []func()
}

func (p *Pending) Add(pr ResponseParams) {
	p.Lock()
	defer p.Unlock()
	log.T.F("adding response hook %s", pr.ID)
	r := &Response{
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

func (p *Pending) FindOlder(t time.Time) (r []*Response) {
	p.Lock()
	defer p.Unlock()
	for i := range p.responses {
		if p.responses[i].Time.Before(t) {
			r = append(r, p.responses[i])
		}
	}
	return
}

func (p *Pending) Find(id nonce.ID) (pr *Response) {
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
func (p *Pending) ProcessAndDelete(id nonce.ID, ifc interface{},
	b slice.Bytes) (found bool, e error) {
	
	p.Lock()
	defer p.Unlock()
	for i := range p.responses {
		if p.responses[i].ID == id {
			log.D.F("deleting response %s", id)
			// Stop the timeout handler.
			p.responses[i].Success.Q()
			for _, fn := range p.responses[i].PostAcct {
				fn()
			}
			e = p.responses[i].Callback(id, ifc, b)
			if i < 1 {
				p.responses = p.responses[1:]
			} else {
				p.responses = append(p.responses[:i],
					p.responses[i+1:]...)
			}
			found = true
			break
		}
	}
	return
}
