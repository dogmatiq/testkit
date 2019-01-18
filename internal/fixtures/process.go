package fixtures

import (
	"github.com/dogmatiq/dogma"
)

// ProcessRoot is a test implementation of dogma.ProcessRoot.
type ProcessRoot struct {
	Value interface{}
}

var _ dogma.ProcessRoot = &ProcessRoot{}
