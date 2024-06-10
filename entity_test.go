package ecs

import (
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestNilEntityMethods(t *testing.T) {
	var e *Entity

	func() {
		defer func() {
			if recover() != nil {
				t.Errorf("TestNilEntityMethods(): Delete is not expected to panic")
			}
		}()
		e.Delete()
	}()
	if want, got := (*Prop)(nil), e.Get(""); want != got {
		t.Errorf("TestNilEntityMethods(): Get = %v, want %v", got, want)
	}
	if want, got := false, e.Has(""); want != got {
		t.Errorf("TestNilEntityMethods(): Has = %v, want %v", got, want)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Errorf("TestNilEntityMethods(): Put is expected to panic")
			}
		}()
		e.Put("")
	}()
	if want, got := (*Prop)(nil), e.Remove(""); want != got {
		t.Errorf("TestNilEntityMethods(): Remove = %v, want %v", got, want)
	}
}

func TestZeroEntity(t *testing.T) {
	var e Entity

	const testKey = "testkey"

	e.Put(testKey) // Put should not panic.
	// Has should return true.
	if want, got := true, e.Has(testKey); want != got {
		t.Errorf("TestZeroEntity(): want Has(%q) = %v, got %v", testKey, want, got)
	}
	e.Get(testKey) // Get should not panic.
	e.Delete()     // Delete should not panic.
}

func TestStubProp(t *testing.T) {
	var e Entity

	if wantProp, gotProp := (*Prop)(nil), e.Get("missing"); wantProp != gotProp {
		t.Errorf("TestStubProp(): got Prop %v, want %v", gotProp, wantProp)
	}
}

func TestRemoveWithSubsequentPut(t *testing.T) {
	const testKey = "testkey"

	var u Universe
	e := u.Entity()

	// Repeated Put and Remove will put the Prop back,
	// allocating a new *Prop for it.
	prop := e.Put(testKey)
	prop.Remove()
	prop2 := e.Put(testKey)

	if got, want := e.Has(testKey), true; got != want {
		t.Errorf("TestRemoveWithSubsequentPut(): got Has() = %v, want %v", got, want)
	}
	if &prop == &prop2 {
		t.Errorf("TestRemoveWithSubsequentPut(): expected Prop not to be reused (%v == %v)", &prop, &prop2)
	}
	gotRange := 0
	u.Range(testKey, func(*Prop) bool { gotRange++; return true })
	if wantRange := 1; gotRange != wantRange {
		t.Errorf("TestRemoveWithSubsequentPut(): got Range visits %d, want %d", gotRange, wantRange)
	}
}

func TestDeleteWithConcurrentPuts(t *testing.T) {
	var u Universe
	e := u.Entity()

	var wg sync.WaitGroup

	// Put 10 thousand Props.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := int64(0); i < 10_000; i++ {
			e.Put(strconv.FormatInt(i, 10))
		}
	}()

	// Delete the Entity.
	wg.Add(1)
	go func() {
		defer wg.Done()
		e.Delete()
	}()

	// Wait.
	wg.Wait()

	// Expect the final number of Props to be 0.
	gotRange := 0
	for i := int64(0); i < 10_000; i++ {
		u.Range(strconv.FormatInt(i, 10), func(*Prop) bool { gotRange++; return true })
	}
	if wantRange := 0; gotRange != wantRange {
		t.Errorf("TestDeleteWithConcurrentPuts(): got Range visits %d, want %d", gotRange, wantRange)
	}
}

func TestEntityExample(t *testing.T) {
	var e Entity

	// Imbue the entity with properties.
	e.Put("sword")
	e.Put("oneHanded")

	// Craft a dynamic name for the item based on held properties.
	var sb strings.Builder
	if e.Has("oneHanded") {
		sb.WriteString("one-handed ")
	} else if e.Has("twoHanded") {
		sb.WriteString("two-handed ")
	}

	if e.Has("sword") {
		sb.WriteString("sword")
	} else {
		sb.WriteString("item")
	}

	if got, want := sb.String(), "one-handed sword"; got != want {
		t.Errorf("TestEntityExample(): got %q, want %q", got, want)
	}
}
