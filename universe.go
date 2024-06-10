package ecs

// Universe provides an isolated scope for Entities and their Props.
//
// An Entity must be assoaicted with a Universe to access the Range method.
//
// Universe is safe for concurrent use.
type Universe struct {
	propDB
}

// Entity creates a new Entity in this Universe.
func (u *Universe) Entity() *Entity { return &Entity{u: u} }
