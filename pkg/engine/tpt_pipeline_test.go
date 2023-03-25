package engine

import (
	"testing"
	
	"github.com/cybriq/qu"
)

func TestNewPipeline(t *testing.T) {
	pA := NewPipeline(1)
	pB := NewPipeline(1)
	pA.Transport = pB
	q := qu.T()
	go func() {
		log.I.F("a->b '%s'", string(<-pB.Receive()))
		q.Q()
	}()
	pA.Send([]byte("testing testing 1 2 3"))
	<-q.Wait()
	q = qu.T()
	go func() {
		log.I.F("b->a '%s'", string(<-pA.Receive()))
		q.Q()
	}()
	pB.Send([]byte("testing testing 1 2 3"))
	<-q.Wait()
}
