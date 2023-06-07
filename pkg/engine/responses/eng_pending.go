package responses

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/qu"
	"github.com/indra-labs/indra/pkg/util/slice"
	"sync"
	"time"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

type (
	Response struct {
		ID       nonce.ID
		SentSize int
		Port     uint16
		Billable []crypto.PubBytes
		Return   crypto.PubBytes
		PostAcct []func()
		sessions.Sessions
		Callback Callback
		Time     time.Time
		Success  qu.C
	}
	ResponseParams struct {
		ID       nonce.ID
		SentSize int
		S        sessions.Sessions
		Billable []crypto.PubBytes
		Ret      crypto.PubBytes
		Port     uint16
		Callback Callback
		PostAcct []func()
	}
	Callback func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error)
	Pending  struct {
		sync.Mutex
		responses []*Response
	}
)

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

func (p *Pending) Find(id nonce.ID) (pr *Response) {
	//p.Lock()
	//defer p.Unlock()
	for i := range p.responses {
		if p.responses[i].ID == id {
			return p.responses[i]
		}
	}
	return
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

// ProcessAndDelete runs the callback and post accounting function list and
// deletes the pending response.
//
// Returns true if it found and deleted a pending response.
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
