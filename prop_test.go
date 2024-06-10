package ecs

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNilPropMethods(t *testing.T) {
	var prop *Prop

	if wantData, gotData := (any)(nil), prop.Data(); wantData != gotData {
		t.Errorf("TestNilPropMethods(): got stub prop Data = %v, want %v", gotData, wantData)
	}
	if wantEntity, gotEntity := (*Entity)(nil), prop.Entity(); wantEntity != gotEntity {
		t.Errorf("TestNilPropMethods(): got stub prop Entity = %v, want %v", gotEntity, wantEntity)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Errorf("TestNilPropMethods(): Key is expected to panic")
			}
		}()
		prop.Key()
	}()
	func() {
		defer func() {
			if recover() == nil {
				t.Errorf("TestNilPropMethods(): PutData is expected to panic")
			}
		}()
		prop.PutData()
	}()
	if wantRemoved, gotRemoved := (*Entity)(nil), prop.Entity(); wantRemoved != gotRemoved {
		t.Errorf("TestNilPropMethods(): got stub prop Removed = %v, want %v", gotRemoved, wantRemoved)
	}
	func() {
		defer func() {
			if recover() != nil {
				t.Errorf("TestNilPropMethods(): Remove is not expected to panic")
			}
		}()
		prop.Remove()
	}()
}

func TestZeroProp(t *testing.T) {
	var p Prop
	if want, got := (*Entity)(nil), p.Entity(); want != got {
		t.Errorf("TestZeroProp(): wanted Entity = %v, got %v", want, got)
	}
	if want, got := "", p.Key(); want != got {
		t.Errorf("TestZeroProp(): wanted Key = %q, got %q", want, got)
	}
	p.PutData(1, 2, 3) // PutData should store the data.
	// Data should return the expected values.
	wantData := []any{1, 2, 3}
	gotData := p.Data() // Data should not panic.
	if diff := cmp.Diff(wantData, gotData); diff != "" {
		t.Errorf("TestZeroProp(): got data diff:\n%s", diff)
	}
	p.Remove() // Remove should not panic.
}
