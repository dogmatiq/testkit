package inflect_test

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	"github.com/dogmatiq/enginekit/message"
	. "github.com/dogmatiq/testkit/internal/inflect"
	"github.com/dogmatiq/testkit/internal/test"
)

func TestInflect(t *testing.T) {
	t.Run("func Sprint()", func(t *testing.T) {
		cases := []struct {
			Template string
			Command  string
			Event    string
			Timeout  string
		}{
			{
				Template: "a <message>",
				Command:  "a command",
				Event:    "an event",
				Timeout:  "a timeout",
			},
			{
				Template: "an <message>",
				Command:  "a command",
				Event:    "an event",
				Timeout:  "a timeout",
			},
			{
				Template: "the <messages>",
				Command:  "the commands",
				Event:    "the events",
				Timeout:  "the timeouts",
			},
			{
				Template: "1 <messages>",
				Command:  "1 command",
				Event:    "1 event",
				Timeout:  "1 timeout",
			},
			{
				Template: "21 <messages>",
				Command:  "21 commands",
				Event:    "21 events",
				Timeout:  "21 timeouts",
			},
			{
				Template: "only 1 <messages>",
				Command:  "only 1 command",
				Event:    "only 1 event",
				Timeout:  "only 1 timeout",
			},
			{
				Template: "only 21 <messages>",
				Command:  "only 21 commands",
				Event:    "only 21 events",
				Timeout:  "only 21 timeouts",
			},
			{
				Template: "<produce> a specific <message>",
				Command:  "execute a specific command",
				Event:    "record a specific event",
				Timeout:  "schedule a specific timeout",
			},
			{
				Template: "the <message> was <produced>",
				Command:  "the command was executed",
				Event:    "the event was recorded",
				Timeout:  "the timeout was scheduled",
			},
			{
				Template: "via a <dispatcher>",
				Command:  "via a dogma.CommandExecutor",
				Event:    "via a dogma.EventRecorder",
				Timeout:  "via a <dispatcher>",
			},
		}

		for _, c := range cases {
			t.Run(c.Template, func(t *testing.T) {
				tests := []struct {
					Kind message.Kind
					Out  string
				}{
					{message.CommandKind, c.Command},
					{message.EventKind, c.Event},
					{message.TimeoutKind, c.Timeout},
				}

				for _, x := range tests {
					t.Run(fmt.Sprintf("%s", x.Kind), func(t *testing.T) {
						test.Expect(
							t,
							"unexpected inflected string",
							Sprint(x.Kind, c.Template),
							x.Out,
						)

						test.Expect(
							t,
							"unexpected uppercase inflected string",
							Sprint(x.Kind, strings.ToUpper(c.Template)),
							strings.ToUpper(x.Out),
						)
					})
				}
			})
		}
	})

	t.Run("func Sprintf()", func(t *testing.T) {
		t.Run("it returns the inflected and substituted string", func(t *testing.T) {
			test.Expect(
				t,
				"unexpected formatted string",
				Sprintf(
					message.CommandKind,
					"the %T <message>",
					CommandA1,
				),
				"the *stubs.CommandStub[github.com/dogmatiq/enginekit/enginetest/stubs.TypeA] command",
			)
		})
	})

	t.Run("func Error()", func(t *testing.T) {
		t.Run("it returns an error with the inflected message", func(t *testing.T) {
			err := Error(
				message.CommandKind,
				"the <message>",
			)

			if err == nil {
				t.Fatal("expected an error")
			}

			test.Expect(t, "unexpected error message", err.Error(), "the command")
		})
	})

	t.Run("func Errorf()", func(t *testing.T) {
		t.Run("it returns an error with the inflected and substituted message", func(t *testing.T) {
			err := Errorf(
				message.CommandKind,
				"the %T <message>",
				CommandA1,
			)

			if err == nil {
				t.Fatal("expected an error")
			}

			test.Expect(
				t,
				"unexpected error message",
				err.Error(),
				"the *stubs.CommandStub[github.com/dogmatiq/enginekit/enginetest/stubs.TypeA] command",
			)
		})
	})
}
