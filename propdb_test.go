package ecs

import "testing"

func TestPropDBCompact(t *testing.T) {
	const testKey = "testkey"

	var u Universe

	e := u.Entity()
	e.Put(testKey)
	e.Remove(testKey) // Remove testKey for e.

	e2 := u.Entity()
	e2.Put(testKey)

	// Fetch the propDB bucket.
	x, ok := u.propDB.data.Load(testKey)
	if !ok {
		t.Fatal("TestPropDBCompact(): expected propDB bucket is missing")
	}
	b := x.(*bucket)

	// Assert the bucket has length 2.
	if gotLen := len(b.data); gotLen != 2 {
		t.Fatalf("TestPropDBCompact(): expected bucket length 1, got %d", gotLen)
	}

	// Do bucketMissesBeforeCompact misses for testKey.
	for i := 0; i < bucketMissesBeforeCompact; i++ {
		u.Range(testKey, func(p *Prop) bool { return true })
	}

	// Assert the bucket is now compacted (is 1).
	if gotLen := len(b.data); gotLen != 1 {
		t.Errorf("TestPropDBCompact(): expected compacted bucket (length 1), got %d", gotLen)
	}
}

func TestRecursiveRange(t *testing.T) {
	// Create a Universe with 2 trees.
	var u Universe
	e := u.Entity()
	e.Put("tree")
	e2 := u.Entity()
	e2.Put("tree")

	rangeVisits := 0

	// Recursive Ranges are allowed even over the same key.
	u.Range("tree", func(p *Prop) bool {
		rangeVisits++
		t.Log(p)
		u.Range("tree", func(p *Prop) bool {
			rangeVisits++
			t.Log(p)
			return true
		})
		return true
	})

	if wantVisits := 6; rangeVisits != wantVisits {
		t.Errorf("TestRecursiveRange(): got Range visits %d, want %d", rangeVisits, wantVisits)
	}
}

func TestRemoveInRange(t *testing.T) {
	// Create a Universe with 1 tree.
	var u Universe
	e := u.Entity()
	e.Put("tree")
	e2 := u.Entity()
	e2.Put("tree")

	rangeVisits := 0

	// Calling Remove inside a Range is always allowed and the
	// Range will not visit the Removed Prop if its deleted.
	// Internally the cleanup of the data is defered until later.
	u.Range("tree", func(p *Prop) bool {
		t.Log(p)
		rangeVisits++
		e.Remove("tree") // Remove both trees.
		e2.Remove("tree")
		return true
	})

	if wantVisits := 1; rangeVisits != wantVisits {
		t.Errorf("TestRemoveInRange(): got Range visits %d, want %d", rangeVisits, wantVisits)
	}
}

func TestPutInRange(t *testing.T) {
	// Create a Universe with 1 tree.
	var u Universe
	e := u.Entity()
	e.Put("tree")

	rangeVisits := 0

	// Calling Put inside a Range is allowed, but new Props
	// will not be visited until the next Range call.
	u.Range("tree", func(p *Prop) bool {
		t.Log(p)
		rangeVisits++
		e2 := u.Entity()
		e2.Put("tree") // Create a new tree.
		return true
	})

	if wantVisits := 1; rangeVisits != wantVisits {
		t.Errorf("TestPutInRange(): got Range visits %d, want %d", rangeVisits, wantVisits)
	}
}
