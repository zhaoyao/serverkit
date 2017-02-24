package buffer

type simpleAlloc struct {
	buf []byte
}

func NewSimpleAlloc(n int) Allocator {
	return &simpleAlloc{}
}

func (d *simpleAlloc) Alloc(n int) (ss []byte) {
	switch {
	case n == 0:
		return []byte{}
	case n >= 512:
		return make([]byte, n)
	default:
		if len(d.buf) < n {
			d.buf = make([]byte, 8192)
		}
		ss, d.buf = d.buf[:n:n], d.buf[n:]
		return ss
	}
}

func (d *simpleAlloc) Free(p []byte) {
}
