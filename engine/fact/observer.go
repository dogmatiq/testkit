package fact

import "sync"

// Observer is an interface that is notified when facts are recorded.
type Observer interface {
	// Notify the observer of a fact.
	Notify(Fact)
}

// ObserverGroup is a collection of observers that can be notified as a group.
type ObserverGroup []Observer

// Notify notifies all of the observers in the set of each of the given fact.
func (s ObserverGroup) Notify(f Fact) {
	for _, o := range s {
		o.Notify(f)
	}
}

// Buffer is an Observer that buffers facts in-memory.
//
// It may be used by multiple goroutines simultaneously.
type Buffer struct {
	m     sync.RWMutex
	facts []Fact
}

// Notify appends f to b.Facts.
func (b *Buffer) Notify(f Fact) {
	b.m.Lock()
	defer b.m.Unlock()

	b.facts = append(b.facts, f)
}

// Facts returns the facts that have been buffered so far.
func (b *Buffer) Facts() []Fact {
	b.m.RLock()
	defer b.m.RUnlock()

	facts := make([]Fact, len(b.facts))
	copy(facts, b.facts)

	return b.facts
}

// Ignore is an observer that ignores fact notifications.
var Ignore ObserverFunc = func(Fact) {}

// ObserverFunc is an adaptor to allow the use of a regular function as an
// observer.
//
// If f is a function with the appropriate signature, ObserverFunc(f) is an
// Observer that calls f.
type ObserverFunc func(Fact)

// Notify calls fn(f).
func (fn ObserverFunc) Notify(f Fact) {
	fn(f)
}
