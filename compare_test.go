package testkit_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	. "github.com/dogmatiq/testkit/internal/fixtures"
	"github.com/dogmatiq/testkit/internal/testingmock"
)

func TestDefaultMessageComparator(t *testing.T) {
	t.Run("the messages are equal", func(t *testing.T) {
		t.Run("plain struct", func(t *testing.T) {
			if !DefaultMessageComparator(CommandA1, CommandA1) {
				t.Fatal("expected messages to be equal")
			}
		})

		t.Run("protocol buffers", func(t *testing.T) {
			if !DefaultMessageComparator(
				NewProtoMessageBuilder().WithValue("<value>").Build(),
				NewProtoMessageBuilder().WithValue("<value>").Build(),
			) {
				t.Fatal("expected messages to be equal")
			}
		})

		t.Run("it ignores unexported fields when comparing protocol buffers messages", func(t *testing.T) {
			a := NewProtoMessageBuilder().WithValue("<value>").Build()
			b := NewProtoMessageBuilder().WithValue("<value>").Build()

			_ = a.String()

			if reflect.DeepEqual(a, b) {
				t.Fatal("unexported fields do not differ")
			}

			if !DefaultMessageComparator(a, b) {
				t.Fatal("expected comparator to ignore unexported fields")
			}
		})
	})

	t.Run("the messages are not equal", func(t *testing.T) {
		t.Run("different types", func(t *testing.T) {
			if DefaultMessageComparator(CommandA1, CommandB1) {
				t.Fatal("expected messages to be different")
			}
		})

		t.Run("protocol buffers", func(t *testing.T) {
			if DefaultMessageComparator(
				NewProtoMessageBuilder().WithValue("<value-a>").Build(),
				NewProtoMessageBuilder().WithValue("<value-b>").Build(),
			) {
				t.Fatal("expected messages to be different")
			}
		})
	})
}

func TestWithMessageComparator(t *testing.T) {
	t.Run("it configures how messages are compared", func(t *testing.T) {
		handler := &IntegrationMessageHandlerStub{
			ConfigureFunc: func(c dogma.IntegrationConfigurer) {
				c.Identity("<handler-name>", "7cb41db6-0116-4d03-80d7-277cc391b47e")
				c.Routes(
					dogma.HandlesCommand[*CommandStub[TypeA]](),
					dogma.RecordsEvent[*EventStub[TypeA]](),
				)
			},
			HandleCommandFunc: func(
				_ context.Context,
				s dogma.IntegrationCommandScope,
				_ dogma.Command,
			) error {
				s.RecordEvent(EventA1)
				return nil
			},
		}

		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "477a9515-8318-4229-8f9d-57d84f463cb7")
				c.Routes(
					dogma.ViaIntegration(handler),
				)
			},
		}

		Begin(
			&testingmock.T{},
			app,
			WithMessageComparator(
				func(a, b dogma.Message) bool {
					return true
				},
			),
		).
			EnableHandlers("<handler-name>").
			Expect(
				ExecuteCommand(CommandA1),
				ToRecordEvent(EventA2), // this would fail without our custom comparator
			)
	})
}
