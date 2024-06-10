package ecs

import "sync/atomic"

// Prop is a property attached to an Entity.
//
// Prop has a Key used to uniquely index it in an Entity
// and Data which carries its payload.
//
// Props add meaning the entities they are attached to.
//
// Example:
//
//	if sword.Has("enchantedFlame") {
//	  if sword.Get("enchantedFlame").Data().(enchantedFlame).RollProc() {
//	    sword.Put("onFire")
//	  }
//	}
//
// Prop is safe for concurrent use.
type Prop struct {
	e        atomic.Pointer[Entity]
	key      string
	attached atomic.Bool
	data     any
}

// Data gets the data from the Prop if any.
func (p *Prop) Data() any {
	if p == nil {
		return nil
	}
	return p.data
}

// PutData sets the data in the Prop.
func (p *Prop) PutData(data ...any) {
	switch len(data) {
	case 0:
		p.data = nil
	case 1:
		p.data = data[0]
	default:
		p.data = data
	}
}

// Key returns the Key for this Prop.
func (p *Prop) Key() string { return p.key }

// Entity returns the Entity containing this Prop or nil if it has none.
//
// Allowing Props to access their Entity enables better code locality and simplicity
// as the Prop update can access the Entity's other Props and make arbitrary changes.
func (p *Prop) Entity() *Entity {
	if p == nil {
		return nil
	}
	return p.e.Load()
}

// Removed returns true after the Prop is removed from its Entity.
func (p *Prop) Removed() bool {
	if p == nil {
		return true
	}
	return !p.attached.Load()
}

// Remove the Prop from its Entity.
func (p *Prop) Remove() {
	if p == nil {
		return
	}
	if e := p.e.Load(); e != nil {
		e.removeKey(p.key)
	}
	p.detatch()
}

func (p *Prop) detatch() { p.attached.Store(false); p.e.Store(nil) }

func (p *Prop) attach(e *Entity) { p.e.Store(e); p.attached.Store(true) }
