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

// Ignore is an observer that ignores facts.
var Ignore ignorer

type ignorer struct{}

// Notify does nothing.
func (ignorer) Notify(Fact) {
}
