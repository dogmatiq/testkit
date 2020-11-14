package testkit

import (
	"github.com/dogmatiq/testkit/assert"
	"github.com/dogmatiq/testkit/engine"
)

// An Expectation is a predicate for determining whether some specific criteria
// was met while performing an action.
type Expectation = assert.Assertion

// ExpectOption is an option that changes the behavior of engine during a call
// to Test.Expect().
type ExpectOption = engine.OperationOption
