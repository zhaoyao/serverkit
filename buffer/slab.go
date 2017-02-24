package buffer

import (
	"github.com/couchbase/go-slab"
)

type slabAlloc struct {
	a *slab.Arena
}

func NewSlabAlloc(startChunkSize int, slabSize int, growthFactor float64) Allocator {
	return &slabAlloc{a: slab.NewArena(startChunkSize, slabSize, growthFactor, nil)}
}

func (a *slabAlloc) Alloc(n int) []byte {
	return a.a.Alloc(n)
}

func (a *slabAlloc) Free(p []byte) {
	if !a.a.DecRef(p) {
		panic("unexpected ref count")
	}
}
