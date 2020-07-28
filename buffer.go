package zgelf

import (
	"fmt"
)

type logBuffer struct {
	buffers   [][]byte
	size      int
}

func NewLogBuffer() *logBuffer {
	return &logBuffer{
		buffers:   make([][]byte, 0),
		size:      0,
	}
}

func (b *logBuffer) Add(data []byte) {
	b.size = b.size + len(data)
	b.buffers = append(b.buffers, data)
}

func (b *logBuffer) Clear() {
	b.size = 0
	b.buffers = make([][]byte, 0)
}

func (b *logBuffer) Copy() *logBuffer {
	c := make([][]byte, len(b.buffers))
	for i, x := range b.buffers {
		c[i] = make([]byte, len(x))
		copy(c[i], x)
	}
	return &logBuffer{
		buffers:   c,
		size:      b.size,
	}
}

func (b *logBuffer) Pull() ([]byte, error) {

	if len(b.buffers) == 0 {
		return nil, fmt.Errorf("no element left in buffer")
	}
	r := b.buffers[:1]
	b.buffers = b.buffers[1:]
	b.size -= len(r[0])
	return r[0], nil
}

func (b *logBuffer) Size() int {
	return b.size
}
