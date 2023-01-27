package ring

import (
	"sync"
	"time"
)

type Load struct {
	Level byte
	time.Time
}

type BufferLoad struct {
	sync.Mutex
	buf     []Load
	lastEMA byte // The time is at the cursor.
	cursor  int
	full    bool
}

// NewBufferLoad creates a new ring buffer of utilization samples.
func NewBufferLoad(size int) *BufferLoad {
	return &BufferLoad{
		buf:    make([]Load, size),
		cursor: -1,
	}
}

func (b *BufferLoad) LastEMA() (out byte) {
	b.Lock()
	defer b.Unlock()
	return b.lastEMA
}

func (b *BufferLoad) Cursor() (out int) {
	b.Lock()
	defer b.Unlock()
	return b.cursor
}

func (b *BufferLoad) Newest() (out Load) {
	b.Lock()
	defer b.Unlock()
	return b.buf[b.cursor]
}

func (b *BufferLoad) Full() (full bool) {
	b.Lock()
	defer b.Unlock()
	return b.full
}

// Get returns the value at the given index or nil if nothing
func (b *BufferLoad) Get(index int) (out byte) {
	b.Lock()
	defer b.Unlock()
	bl := len(b.buf)
	if index < bl {
		cursor := b.cursor + index
		if cursor > bl {
			cursor = cursor - bl
		}
		return b.buf[cursor].Level
	}
	return
}

// Len returns the length of the buffer, which grows until it fills, after which
// this will always return the size of the buffer
func (b *BufferLoad) Len() (length int) {
	b.Lock()
	defer b.Unlock()
	if b.full {
		return len(b.buf)
	}
	return b.cursor
}

// Add a new value to the cursor position of the ring buffer
func (b *BufferLoad) Add(value byte) {
	b.Lock()
	defer b.Unlock()
	b.cursor++
	if b.cursor == len(b.buf) {
		b.cursor = 0
		if !b.full {
			b.full = true
		}
	}
	b.buf[b.cursor].Level = value
	b.buf[b.cursor].Time = time.Now()
}

// ForEach is an iterator that can be used to process every element in the
// buffer
func (b *BufferLoad) ForEach(fn func(v Load) error) (e error) {
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
