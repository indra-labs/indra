package ring

import (
	"sync"
	"time"
)

type Latency struct {
	time.Duration
	time.Time
}

type BufferLatency struct {
	sync.Mutex
	buf     []Latency
	lastEMA time.Duration // The time is the current cursor.
	cursor  int
	full    bool
}

// NewBufferLatency creates a new ring buffer of latency samples.
func NewBufferLatency(size int) *BufferLatency {
	return &BufferLatency{
		buf:    make([]Latency, size),
		cursor: -1,
	}
}

func (b *BufferLatency) LastEMA() (out time.Duration) {
	b.Lock()
	defer b.Unlock()
	return b.lastEMA
}

func (b *BufferLatency) Cursor() (out int) {
	b.Lock()
	defer b.Unlock()
	return b.cursor
}

func (b *BufferLatency) Newest() (out Latency) {
	b.Lock()
	defer b.Unlock()
	return b.buf[b.cursor]
}

func (b *BufferLatency) LastTime() time.Time {
	b.Lock()
	defer b.Unlock()
	return b.buf[b.cursor].Time
}

func (b *BufferLatency) Full() (full bool) {
	b.Lock()
	defer b.Unlock()
	return b.full
}

// Get returns the value at the given index or nil if nothing
func (b *BufferLatency) Get(index int) (out time.Duration) {
	b.Lock()
	defer b.Unlock()
	bl := len(b.buf)
	if index < bl {
		cursor := b.cursor + index
		if cursor > bl {
			cursor = cursor - bl
		}
		return b.buf[cursor].Duration
	}
	return
}

// Len returns the length of the buffer, which grows until it fills, after which
// this will always return the size of the buffer
func (b *BufferLatency) Len() (length int) {
	b.Lock()
	defer b.Unlock()
	if b.full {
		return len(b.buf)
	}
	return b.cursor
}

// Add a new value to the cursor position of the ring buffer
func (b *BufferLatency) Add(value time.Duration) {
	b.Lock()
	defer b.Unlock()
	b.cursor++
	if b.cursor == len(b.buf) {
		b.cursor = 0
		if !b.full {
			b.full = true
		}
	}
	b.buf[b.cursor].Duration = value
	b.buf[b.cursor].Time = time.Now()
}

// ForEach is an iterator that can be used to process every element in the
// buffer
func (b *BufferLatency) ForEach(fn func(v Latency) error) (e error) {
	b.Lock()
	defer b.Unlock()
	c := b.cursor
	i := c + 1
	if i == len(b.buf) {
		// hit the end
		i = 0
	}
	if !b.full {
		// buffer not yet full
		i = 0
	}
	for ; ; i++ {
		if i == len(b.buf) {
			// passed the end
			i = 0
		}
		if i == c {
			// reached cursor again
			break
		}
		if e = fn(b.buf[i]); e != nil {
			break
		}
	}
	return
}
