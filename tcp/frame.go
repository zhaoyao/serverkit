package tcp

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
)

type BufferAllocFunc func(n int) []byte

var defaultBufferAlloc = func(n int) []byte {
	return make([]byte, n)
}

func EncodeFrame4BE(w io.Writer, p []byte) (int64, error) {
	if err := binary.Write(w, binary.BigEndian, uint32(len(p))); err != nil {
		return 0, err
	}

	n, err := w.Write(p)
	return int64(n) + 4, err
}

func DecodeFrame4BE(r io.Reader, f BufferAllocFunc) ([]byte, error) {
	if f == nil {
		f = defaultBufferAlloc
	}
	var plen uint32
	if err := binary.Read(r, binary.BigEndian, &plen); err != nil {
		return nil, err
	}

	p := f(int(plen))
	if _, err := io.ReadAtLeast(r, p, int(plen)); err != nil {
		return nil, err
	}

	return p, nil
}

type OutputFrame struct {
	packed bool
	mr     io.Reader
	hdr    []byte
	buf    *bytes.Buffer
}

func (f *OutputFrame) Write(p []byte) (int, error) {
	return f.buf.Write(p)
}

func (f *OutputFrame) pack() {
	binary.BigEndian.PutUint32(f.hdr, uint32(f.buf.Len()))
	f.mr = io.MultiReader(bytes.NewReader(f.hdr), f.buf)
	f.packed = true
}

func (f *OutputFrame) Read(p []byte) (int, error) {
	if !f.packed {
		f.pack()
	}
	return f.mr.Read(p)
}

func (f *OutputFrame) WriteTo(w io.Writer) (int64, error) {
	return EncodeFrame4BE(w, f.buf.Bytes())
}
