package ecs

import (
	"slices"
	"sync"
	"sync/atomic"
)

// bucketMissesBeforeCompact is the number of misses required to compact the bucket.
const bucketMissesBeforeCompact = 500

// propDB maintains a database of all Props.
type propDB struct {
	data sync.Map // string -> *bucket
}

// Range over all Props in the Universe with a matching key and call fn.
// The Range stops if fn returns false.
//
// Range is safe for concurrent use.
// Range does not block and fn may call any method on the Universe during iteration
// including putting and removing Props, and even recursive Range calls.
func (d *propDB) Range(key string, fn func(*Prop) bool) {
	x, loaded := d.data.Load(key)
	if !loaded {
		return
	}
	b := x.(*bucket)
	b.rangeProps(fn)
	b.tryCompact()
}

func (d *propDB) getOrCreateBucket(key string) *bucket {
	e, loaded := d.data.Load(key)
	if !loaded {
		e, _ = d.data.LoadOrStore(key, new(bucket))
	}
	return e.(*bucket)
}

func (d *propDB) append(key string, prop *Prop) {
	b := d.getOrCreateBucket(key)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data = append(b.data, prop)
}

type bucket struct {
	data   []*Prop
	mu     sync.RWMutex
	misses atomic.Int64
}

func (b *bucket) rangeProps(fn func(*Prop) bool) {
	b.mu.RLock()
	// Read the slice header inside the RLock.
	// If the bucket is appended in the Range,
	// we won't visit the new elements this time.
	data := b.data
	b.mu.RUnlock()
	for _, e := range data {
		if e.Removed() {
			b.misses.Add(1)
		} else if !fn(e) {
			break
		}
	}
}

// tryCompact compacts the bucket if we have enough misses
// by removing detatched Props.
func (b *bucket) tryCompact() {
	// Compact the bucket if we have enough misses.
	if b.misses.Load() < bucketMissesBeforeCompact {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	begin := 0 // Scan to the first unremoved Prop.
	for ; begin < len(b.data) && b.data[begin].Removed(); begin++ {
	}
	// Remove the Props by compacting the bucket.
	b.data = slices.CompactFunc(b.data[begin:], func(a, b *Prop) bool { return b.Removed() })
	b.misses.Store(0)
}
