package handler

import "fmt"

// EmptyInstanceIDError indicates that an aggregate or process message handler has
// attempted to route a message to an instance with an empty ID.
type EmptyInstanceIDError struct {
	HandlerName string
	HandlerType Type
}

func (e EmptyInstanceIDError) Error() string {
	return fmt.Sprintf(
		"the '%s' %s message handler attempted to route a message to an empty instance ID",
		e.HandlerName,
		e.HandlerType,
	)
}

// NilRootError indicates that an aggregate or process message handler has
// returned a nil "root" value from its New() method.
type NilRootError struct {
	HandlerName string
	HandlerType Type
}

func (e NilRootError) Error() string {
	return fmt.Sprintf(
		"the '%s' %s message handler produced a nil root",
		e.HandlerName,
		e.HandlerType,
	)
}

// EventNotRecordedError indicates that an aggregate instance was created
// or destroyed without recording an event.
type EventNotRecordedError struct {
	HandlerName  string
	InstanceID   string
	WasDestroyed bool
}

func (e EventNotRecordedError) Error() string {
	s := "created"

	if e.WasDestroyed {
		s = "destroyed"
	}

	return fmt.Sprintf(
		"the '%s' aggregate message handler %s the '%s' instance without recording an event",
		e.HandlerName,
		s,
		e.InstanceID,
	)
}
