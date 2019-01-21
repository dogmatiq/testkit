package fact

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
type Buffer struct {
	Facts []Fact
}

// Notify appends f to b.Facts.
func (b *Buffer) Notify(f Fact) {
	b.Facts = append(b.Facts, f)
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
