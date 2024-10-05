package inflect

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/dogmatiq/enginekit/message"
)

var substitutions = map[message.Kind]map[string]string{
	message.CommandKind: {
		"<message>":    "command",
		"<messages>":   "commands",
		"<produce>":    "execute",
		"<produced>":   "executed",
		"<producing>":  "executing",
		"<dispatcher>": "dogma.CommandExecutor",
	},
	message.EventKind: {
		"<message>":    "event",
		"<messages>":   "events",
		"<produce>":    "record",
		"<produced>":   "recorded",
		"<producing>":  "recording",
		"<dispatcher>": "dogma.EventRecorder",
	},
	message.TimeoutKind: {
		"<message>":   "timeout",
		"<messages>":  "timeouts",
		"<produce>":   "schedule",
		"<produced>":  "scheduled",
		"<producing>": "scheduling",
	},
}

var corrections = map[string]string{
	"an command": "a command",
	"a event":    "an event",
	"an timeout": "a timeout",
	"1 commands": "1 command",
	"1 events":   "1 event",
	"1 timeouts": "1 timeout",
}

// Sprint formats a string, inflecting words in s match the message kind k.
func Sprint(k message.Kind, s string) string {
	for k, v := range substitutions[k] {
		s = strings.ReplaceAll(s, k, v)
		s = strings.ReplaceAll(s, strings.ToUpper(k), strings.ToUpper(v))
	}

	for k, v := range corrections {
		s = replaceAll(s, k, v)
		s = replaceAll(s, strings.ToUpper(k), strings.ToUpper(v))
	}

	return s
}

// Sprintf formats a string, inflecting words in f match the message kind k.
func Sprintf(k message.Kind, f string, v ...any) string {
	return Sprint(
		k,
		fmt.Sprintf(f, v...),
	)
}

// Error returns a new error, inflecting words in s to match the message kind k.
func Error(k message.Kind, s string) error {
	return errors.New(Sprint(k, s))
}

// Errorf returns a new error, inflecting words in f to match the message kind k.
func Errorf(k message.Kind, f string, v ...any) error {
	return errors.New(Sprintf(k, f, v...))
}

// replaceAll replaces all instances of k with v, at word boundaries.
func replaceAll(s, k, v string) string {
	return regexp.MustCompile(`(?m)\b`+regexp.QuoteMeta(k)+`\b`).ReplaceAllLiteralString(s, v)
}
