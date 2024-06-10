# go-ecs

Compositional entity-component system for games.

## Import

```go
import (
  "github.com/wenooij/go-ecs"
)
```

## Usage

### Create a Universe

This is like a namespace for Entities.

```go
var u ecs.Universe
```

Entities can be created directly from the `Universe`.

```go
e := u.Entity() // New empty Entity in this Universe.
```

Entities cannot be moved into a `Universe` later.

Using an `Entity` without a Universe is allowed but, as we'll see, `Range` is not supported.

### Create some `Prop`s

```go
type enchantedFlame struct { level int }
type onFire struct { timeLeft int }

func (x enchantedFlame) RollProc() bool { /* ... */ }
```

### Create an `Entity` with some Props

```go
e := u.Entity()
e.Put("sword")
e.Put("oneHanded")
e.Put("enchantedFlame", enchantedFlame{2})
```

### Write code using the `Entity`

```go
if e.Has("enchantedFlame") {
  if e.Get("enchantedFlame").Data().(enchantedFlame).RollProc() {
    e.Put("onFire")
  }
}
```

### Update all Props efficiently in batch

Updating properties this way is much more efficient because the tight `Range` loop helps with temporal and spacial locality, which helps with caching. The alternative, updating entities directly, interleves unrelated operations, and results in more cache misses. 

```go
func updateVelocities(u *ecs.Universe) {
  u.Range("vel", func(p *ecs.Prop) {
    p.Data().(*Vec).Add(p.Entity().Get("acc").(*Vec))
  })
}

func updatePositions(u *ecs.Universe) {
  u.Range("pos", func(p *ecs.Prop) {
    p.Data().(*Vec).Add(p.Entity().Get("vel").(*Vec))
  })
}

func updateHealthBars(u *ecs.Universe) {
  u.Range("health", func(p *ecs.Prop) bool {
    if p.Data.(enemyHealth).health <= 0 {
      p.Entity().Delete() // Die!
    }
  })
}
```

## Tips and Tricks

### Key-only Prop

Create a Prop that doesn't have any data.

```go
e.Put("humanoid")
```

### Lazy-initialization

A Key-only prop can be expanded later on initialization.

```go
e.Put("enemy")

// ...

func spawnEnemies(u *ecs.Universe) {
  u.Range("enemy", func(p *ecs.Prop) bool {
    if !p.Entity().Has("location") {
      p.Entity().Put("location", newSpawnLocation())
    }
  })
}
```

### Timer

```go
e.Put("poison")

// ...

func updatePoison(u *ecs.Universe) {
  u.Range("poison", func(p *ecs.Prop) bool {
    switch {
    case p.Data() == nil: // Initialize poison.
      p.PutData(Tick(10))
    default:
      tick := p.Data().(Tick)-1
      if p.PutData(tick); tick <= 0 {
        p.Delete()
      }
    }
  })
}
```

### State Machines

State machines can be used to control behavior with state transitions and two or more Props.

A simple state machine can be encoded in a switch-case statement.

```go
func updateStates(u *ecs.Universe) {
  u.Range("state", func(p *ecs.Prop) bool {
    switch p.Data() {
    case nil: // Initial state.
      p.PutData(s0)
    case s0:
      if condition1Met() {
        p.PutData(s1)
      }
    case s1:
      if condition2Met() {
        p.PutData(s2)
      }
    // ...
    }
  })
}
```

More complex machines can use the "online" state machine pattern.

### Online State Machine

```go
type State func(*ecs.Entity) State

func state0(*ecs.Entity) State {
  if condition1Met() {
    return state1
  }
}

func state1(*ecs.Entity) { /* ... */ }

// ... Add more state implementations here...

func updateStates(u *ecs.Universe) {
  u.Range("state", func(p *ecs.Prop) bool {
    p.PutData(p.Data.(State)(p.Entity()))
    return true
  })
}
```

See the test files for more examples.
