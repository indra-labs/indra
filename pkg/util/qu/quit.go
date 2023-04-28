package qu

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
	
	"go.uber.org/atomic"
	
	"git-indra.lan/indra-labs/indra"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

// C is your basic empty struct signalling channel
type C chan struct{}

var (
	createdList     []string
	createdChannels []C
	mx              sync.Mutex
	logEnabled      atomic.Bool
	KillAll         = T()
)

// SetLogging switches on and off the channel logging
func SetLogging(on bool) {
	logEnabled.Store(on)
}

func l(a ...interface{}) {
	if logEnabled.Load() {
		log.D.Ln(a...)
	}
}

// T creates an unbuffered chan struct{} for trigger and quit signalling (momentary and breaker switches)
func T() C {
	mx.Lock()
	defer mx.Unlock()
	msg := Caller("chan from", 1)
	l("created", msg)
	createdList = append(createdList, msg)
	o := make(C)
	createdChannels = append(createdChannels, o)
	return o
}

// Ts creates a buffered chan struct{} which is specifically intended for signalling without blocking, generally one is
// the size of buffer to be used, though there might be conceivable cases where the channel should accept more signals
// without blocking the caller
func Ts(n int) C {
	mx.Lock()
	defer mx.Unlock()
	msg := Caller("buffered chan from", 1)
	l("created", msg)
	createdList = append(createdList, msg)
	o := make(C, n)
	createdChannels = append(createdChannels, o)
	return o
}

// Q closes the channel, which makes it emit a nil every time it is selected
func (c C) Q() {
	l(
		func() (o string) {
			loc := getLocForChan(c)
			mx.Lock()
			defer mx.Unlock()
			if !testChanIsClosed(c) {
				close(c)
				return "closing chan from " + loc + Caller(
					"\n"+strings.Repeat(
						" ", 48,
					)+"from", 1,
				)
			} else {
				return "from" + Caller("", 1) + "\n" + strings.Repeat(
					" ", 48,
				) +
					"channel " + loc + " was already closed"
			}
		}(),
	)
}

// Signal sends struct{}{} on the channel which functions as a momentary switch, useful in pairs for stop/start
func (c C) Signal() {
	l(func() (o string) { return "signalling " + getLocForChan(c) }())
	c <- struct{}{}
}

// Wait should be placed with a `<-` in a select case in addition to the channel variable name
func (c C) Wait() <-chan struct{} {
	l(
		func() (o string) {
			return "waiting on " + getLocForChan(c) + Caller(
				"at", 1,
			)
		}(),
	)
	return c
}

// testChanIsClosed allows you to see whether the channel has been closed so you can avoid a panic by trying to close or
// signal on it
func testChanIsClosed(ch C) (o bool) {
	if ch == nil {
		return true
	}
	select {
	case <-ch:
		o = true
	default:
	}
	return
}

// getLocForChan finds which record connects to the channel in question
func getLocForChan(c C) (s string) {
	s = "not found"
	mx.Lock()
	for i := range createdList {
		if i >= len(createdChannels) {
			break
		}
		if createdChannels[i] == c {
			s = createdList[i]
		}
	}
	mx.Unlock()
	return
}

// once a minute clean up the channel cache to remove closed channels no longer in use
func init() {
	go func() {
	out:
		for {
			select {
			case <-time.After(time.Minute):
				log.D.Ln("cleaning up closed channels")
				var c []C
				var ll []string
				mx.Lock()
				for i := range createdChannels {
					if i >= len(createdList) {
						break
					}
					if testChanIsClosed(createdChannels[i]) {
					} else {
						c = append(c, createdChannels[i])
						ll = append(ll, createdList[i])
					}
				}
				createdChannels = c
				createdList = ll
				mx.Unlock()
			case <-KillAll.Wait():
				break out
			}
		}
	}()
}

// PrintChanState creates an output showing the current state of the channels being monitored
// This is a function for use by the programmer while debugging
func PrintChanState() {
	mx.Lock()
	for i := range createdChannels {
		if i >= len(createdList) {
			break
		}
		if testChanIsClosed(createdChannels[i]) {
			log.D.Ln(">>> closed", createdList[i])
		} else {
			log.D.Ln("<<< open", createdList[i])
		}
	}
	mx.Unlock()
}

// GetOpenChanCount returns the number of qu channels that are still open
// todo: this needs to only apply to unbuffered type
func GetOpenChanCount() (o int) {
	mx.Lock()
	var c int
	for i := range createdChannels {
		if i >= len(createdChannels) {
			break
		}
		if testChanIsClosed(createdChannels[i]) {
			c++
		} else {
			o++
		}
	}
	mx.Unlock()
	return
}

func Caller(comment string, skip int) string {
	_, file, line, _ := runtime.Caller(skip + 1)
	o := fmt.Sprintf("%s: %s:%d", comment, file, line)
	return o
}
