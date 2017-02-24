package buffer

type Allocator interface {
	Alloc(n int) []byte
	Free(p []byte)
}
