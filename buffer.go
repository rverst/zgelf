package zgelf

type logBuffer struct {
	buffers   [][]byte
	size      int
	compress  bool
	delimiter []byte
}

func NewLogBuffer() *logBuffer {
	return &logBuffer{
		buffers:   make([][]byte, 0),
		size:      0,
		compress:  false,
		delimiter: make([]byte, 0),
	}
}

func (b *logBuffer) Add(data []byte) {
	b.size = b.size + len(data)
	b.buffers = append(b.buffers, data)
}

func (b *logBuffer) EnableCompression() {
	b.compress = true
}

func (b *logBuffer) Clear() {
	b.size = 0
	b.buffers = make([][]byte, 0)
	b.delimiter = make([]byte, 0)
	b.compress = false
}

func (b *logBuffer) Copy() *logBuffer {
	c := make([][]byte, len(b.buffers))
	for i, x := range b.buffers {
		c[i] = make([]byte, len(x))
		copy(c[i], x)
	}
	d := make([]byte, len(b.delimiter))
	copy(d, b.delimiter)
	return &logBuffer{
		buffers:   c,
		delimiter: d,
		size:      b.size,
		compress:  b.compress,
	}
}

func (b *logBuffer) SetDelimiter(d []byte) {
	b.delimiter = d
}

func (b *logBuffer) Size() int {
	return b.size
}

func (b *logBuffer) Read(p []byte) (n int, err error) {

	p = make([]byte, 0)
	n = 0
	for _, x := range b.buffers {
		p = append(p, x...)
		n += len(x)
		if len(b.delimiter) > 0 {
			p = append(p, b.delimiter...)
			n += len(x)
		}
	}
	return n, nil
}
