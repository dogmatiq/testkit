package assert

import (
	"fmt"
	"strings"

	"github.com/dogmatiq/configkit/message"
)

// inflect formats a string, inflecting words to suit the given message role r.
//
// For example, if r is message.EventRole, the generic string "<produced>" is
// replaced with the event-specific terminology "recorded".
func inflect(r message.Role, f string, v ...interface{}) string {
	r.MustBe(message.CommandRole, message.EventRole)

	if r == message.CommandRole {
		f = strings.ReplaceAll(f, "<message>", "command")
		f = strings.ReplaceAll(f, "<messages>", "commands")
		f = strings.ReplaceAll(f, "<produce>", "execute")
		f = strings.ReplaceAll(f, "<produced>", "executed")
	} else {
		f = strings.ReplaceAll(f, "<message>", "event")
		f = strings.ReplaceAll(f, "<messages>", "events")
		f = strings.ReplaceAll(f, "<produce>", "record")
		f = strings.ReplaceAll(f, "<produced>", "recorded")
	}

	f = strings.ReplaceAll(f, "an command", "a command")
	f = strings.ReplaceAll(f, "a event", "an event")

	return fmt.Sprintf(f, v...)
}
