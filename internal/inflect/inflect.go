package inflect

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dogmatiq/configkit/message"
)

// Sprint formats a string, inflecting words in s match the message role r.
func Sprint(r message.Role, s string) string {
	switch r {
	case message.CommandRole:
		s = strings.ReplaceAll(s, "<message>", "command")
		s = strings.ReplaceAll(s, "<messages>", "commands")
		s = strings.ReplaceAll(s, "<produce>", "execute")
		s = strings.ReplaceAll(s, "<produced>", "executed")
		s = strings.ReplaceAll(s, "<dispatcher>", "dogma.CommandExecutor")
	case message.EventRole:
		s = strings.ReplaceAll(s, "<message>", "event")
		s = strings.ReplaceAll(s, "<messages>", "events")
		s = strings.ReplaceAll(s, "<produce>", "record")
		s = strings.ReplaceAll(s, "<produced>", "recorded")
		s = strings.ReplaceAll(s, "<dispatcher>", "dogma.EventRecorder")
	case message.TimeoutRole:
		s = strings.ReplaceAll(s, "<message>", "timeout")
		s = strings.ReplaceAll(s, "<messages>", "timeouts")
		s = strings.ReplaceAll(s, "<produce>", "schedule")
		s = strings.ReplaceAll(s, "<produced>", "scheduled")
	}

	s = strings.ReplaceAll(s, "an command", "a command")
	s = strings.ReplaceAll(s, "a event", "an event")
	s = strings.ReplaceAll(s, "an timeout", "a timeout")

	return s
}

// Sprintf formats a string, inflecting words in f match the message role r.
func Sprintf(r message.Role, f string, v ...interface{}) string {
	return fmt.Sprintf(
		Sprint(r, f),
		v...,
	)
}

// Error returns a new error, inflecting words in s to match the message role r.
func Error(r message.Role, s string) error {
	return errors.New(Sprint(r, s))
}

// Errorf returns a new error, inflecting words in f to match the message role r.
func Errorf(r message.Role, f string, v ...interface{}) error {
	return fmt.Errorf(
		Sprint(r, f),
		v...,
	)
}
