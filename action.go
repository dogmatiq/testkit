package testkit

// Action is an interface for any action that can be performed within a test.
//
// Actions always attempt to cause some state change within the engine or
// application.
type Action interface {
	// Heading returns a human-readable description of the action, used as a
	// heading within the test report.
	//
	// Any engine activity as a result of this action is logged beneath this
	// heading.
	Heading() string

	// Apply performs the action within the context of a specific test.
	Apply(
		t *Test,
		e Expectation,
		o ExpectOptionSet,
	)
}
