package assert

import (
	"fmt"
	"strings"

	"github.com/dogmatiq/enginekit/message"
)

func inflect(r message.Role, f string, v ...interface{}) string {
	r.MustBe(message.CommandRole, message.EventRole)

	if r == message.CommandRole {
		f = strings.Replace(f, "<message>", "command", -1)
		f = strings.Replace(f, "<messages>", "commands", -1)
		f = strings.Replace(f, "<produce>", "execute", -1)
		f = strings.Replace(f, "<produced>", "executed", -1)
	} else {
		f = strings.Replace(f, "<message>", "event", -1)
		f = strings.Replace(f, "<messages>", "events", -1)
		f = strings.Replace(f, "<produce>", "record", -1)
		f = strings.Replace(f, "<produced>", "recorded", -1)
	}

	f = strings.Replace(f, "an command", "a command", -1)
	f = strings.Replace(f, "a event", "an event", -1)

	return fmt.Sprintf(f, v...)
}
