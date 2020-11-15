package testkit

import (
	"github.com/dogmatiq/testkit/assert"
)

// An Expectation is a predicate for determining whether some specific criteria
// was met while performing an action.
type Expectation = assert.Assertion

// ExpectOptionSet is a set of options that dictate the behavior of the
// Test.Expect() method.
type ExpectOptionSet = assert.ExpectOptionSet

// ExpectOption is an option that changes the behavior the Test.Expect() method.
type ExpectOption func(*ExpectOptionSet)
