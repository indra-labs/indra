package ring

import (
	"sync"
	"time"
)

type BufferFailure struct {
	sync.Mutex
	buf     []time.Time
	lastEMA time.Duration // The time is the current cursor.
	cursor  int
	full    bool
}

// NewBufferFailure creates a new ring buffer of the history of failures.
func NewBufferFailure(size int) *BufferFailure {
	return &BufferFailure{
		buf:    make([]time.Time, size),
		cursor: -1,
	}
}

func (b *BufferFailure) LastEMA() (out time.Duration) {
	b.Lock()
	defer b.Unlock()
	return b.lastEMA
}

func (b *BufferFailure) Cursor() (out int) {
	b.Lock()
	defer b.Unlock()
	return b.cursor
}

func (b *BufferFailure) Newest() (out time.Time) {
	b.Lock()
	defer b.Unlock()
	return b.buf[b.cursor]
}

func (b *BufferFailure) Full() (full bool) {
	b.Lock()
	defer b.Unlock()
	return b.full
}

// Get returns the value at the given index or nil if nothing
func (b *BufferFailure) Get(index int) (out time.Time) {
	b.Lock()
	defer b.Unlock()
	bl := len(b.buf)
	if index < bl {
		cursor := b.cursor + index
		if cursor > bl {
			cursor = cursor - bl
		}
		return b.buf[cursor]
	}
	return
}

// Len returns the length of the buffer, which grows until it fills, after which
// this will always return the size of the buffer
func (b *BufferFailure) Len() (length int) {
	b.Lock()
	defer b.Unlock()
	if b.full {
		return len(b.buf)
	}
	return b.cursor
}

// Add a new value to the cursor position of the ring buffer
func (b *BufferFailure) Add(value time.Time) {
	b.Lock()
	defer b.Unlock()
	b.cursor++
	if b.cursor == len(b.buf) {
		b.cursor = 0
		if !b.full {
			b.full = true
		}
	}
	b.buf[b.cursor] = value
}

// ForEach is an iterator that can be used to process every element in the
// buffer
func (b *BufferFailure) ForEach(fn func(v time.Time) error) (e error) {
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
			// passed the end"
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
