package fixtures

import "github.com/dogmatiq/dogma"

// AggregateRoot is a test implementation of dogma.AggregateRoot.
type AggregateRoot struct {
	Value interface{}
}

var _ dogma.AggregateRoot = &AggregateRoot{}

// ApplyEvent updates the aggregate instance to reflect the fact that a
// particular domain event has occurred.
func (v *AggregateRoot) ApplyEvent(dogma.Message) {
	panic("not implemented")
}
