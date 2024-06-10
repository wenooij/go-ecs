// Package ecs defines the Entity-component system primitives Entity, Prop, and Universe.
//
// The crux of this package is the ability to define Entities, put Props on them,
// and implement game engines which call Universe.Range to update all Props in batch.
// In order to take advantage of the Range method, we should have a single Universe
// for active Entities and use Universe.Entity to create new ones.
//
// All types are safe for concurrent use.
package ecs

import (
	"sync"
)

// Entity is a collection of component Props.
//
// An entity has no implicit meaning other than those
// granted by the Props it contains.
//
// An Entity is associated with a Universe if constructed
// through the Universe.Entity method.
//
// Entity is safe for concurrent use.
type Entity struct {
	u        *Universe
	props    sync.Map // string -> *Prop
	deleted  bool
	deleteMu sync.RWMutex
}

// Has returns true if the Entity contains the Prop.
//
// Consider only using Has for key Props when there's
// no data to fetch. Example:
//
//	if tool.Has("damaged") {
//		println("Your tool is too damaged to use.")
//	}
//
// Use Get otherwise.
func (e *Entity) Has(key string) bool { return e.Get(key) != nil }

// Get returns the requested Prop or nil if none exists.
func (e *Entity) Get(key string) *Prop {
	e.deleteMu.RLock()
	defer e.deleteMu.RUnlock()
	return e.loadProp(key)
}

// locks excluded: deleteMu.
func (e *Entity) loadProp(key string) *Prop {
	if e == nil || e.deleted {
		return nil
	}
	if x, ok := e.props.Load(key); ok && !x.(*Prop).Removed() {
		return x.(*Prop)
	}
	return nil
}

// Put a Prop on this Entity overwriting old data and return the Prop.
// Put may return nil if the Entity has been Deleted.
//
// The variadic args are a convinence allowing one to specifiy
// Props with no data or multiple data entries (see Prop.PutData
// for exact semantics).
//
// If the Entity is associated with a Universe, it will appear in the
// next Range call.
func (e *Entity) Put(key string, data ...any) (prop *Prop) {
	if e == nil {
		panic("Put called on nil Entity")
	}
	e.deleteMu.RLock()
	defer e.deleteMu.RUnlock()
	if e.deleted {
		return nil
	}
	defer func() { prop.PutData(data...) }()
	if prop := e.loadProp(key); prop != nil {
		return prop
	}
	newProp := &Prop{key: key}
	newProp.attach(e)
	if x, loaded := e.props.LoadOrStore(key, newProp); loaded {
		// We don't reuse Removed Props because those will eventually
		// be GCed from the propDB which could lead to a data race.
		// Instead, if Prop was Removed Store newProp.
		if prop := x.(*Prop); !prop.Removed() {
			return prop
		}
		e.props.Store(key, newProp)
	}
	if e.u != nil {
		e.u.append(key, newProp)
	}
	return newProp
}

// Remove the Prop from the Entity and return the removed Prop if it existed.
func (e *Entity) Remove(key string) (removed *Prop) {
	if e == nil || e.deleted {
		return nil
	}
	e.deleteMu.RLock()
	defer e.deleteMu.RUnlock()
	x, loaded := e.props.LoadAndDelete(key)
	if !loaded {
		return nil
	}
	prop := x.(*Prop)
	prop.detatch()
	return prop
}

func (e *Entity) removeKey(key string) { e.props.Delete(key) }

// Delete deletes the Entity by calling Delete on all its Props
// and removing it from the Universe.
//
// Example:
//
//	var u Universe
//	e := u.Entity()
//	jello := e.Put("jello")
//	fire := e.Put("fire")
//	e.Delete()
//	// jello and fire are now Removed
//	// and no longer show up in u.Range.
func (e *Entity) Delete() {
	if e == nil {
		return
	}
	e.deleteMu.Lock()
	defer e.deleteMu.Unlock()
	// After holding deleteMu's lock, the map cannot be modified.
	// Therefore the following Range is guaranteed to detatch all Props.
	// Also, setting deleted below guarantees no mutation after we release deleteMu.
	e.deleted = true // Mark the Entity Deleted.
	// Detatch all Props.
	e.props.Range(func(_, value any) bool { value.(*Prop).detatch(); return true })
	e.props = sync.Map{} // Map data may now be GCed.
	e.u = nil            // Unlink the Universe.
}
