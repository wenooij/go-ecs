package ecs

import "testing"

func TestRangeDeleteAndRemoveEntityProps(t *testing.T) {
	var u Universe

	e := u.Entity()
	e.Put("brown")
	e.Put("shoe")

	e2 := u.Entity()
	e2.Put("brown")
	e2.Put("shoe")
	e2.Delete() // e is Deleted.

	e3 := u.Entity()
	e3.Put("brown")
	e3.Put("shoe")
	e3.Remove("shoe") // e3 shoe is Removed.

	// Range over the shoes.
	// We expect only 1 undeleted shoe.
	gotShoes := 0
	u.Range("shoe", func(p *Prop) bool { gotShoes++; return true })

	if gotShoes != 1 {
		t.Errorf("TestRangeDeleteAndRemoveEntityProps(): got %d shoe Props, want 2", gotShoes)
	}

	// Range over "brown".
	// We expect 2.
	gotBrown := 0
	u.Range("brown", func(p *Prop) bool { gotBrown++; return true })

	if gotBrown != 2 {
		t.Errorf("TestRangeDeleteAndRemoveEntityProps(): got %d \"brown\" Props, want 2", gotBrown)
	}
}
