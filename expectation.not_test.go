package testkit_test

import (
	"context"
	"testing"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	"github.com/dogmatiq/testkit/x/xtesting"
)

func TestNot(t *testing.T) {
	newFixture := func() (*testingmock.T, *Test) {
		mt := &testingmock.T{FailSilently: true}
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "00df8612-2fd4-4ae3-9acf-afc2b4daf272")
				c.Routes(
					dogma.ViaIntegration(&IntegrationMessageHandlerStub{
						ConfigureFunc: func(c dogma.IntegrationConfigurer) {
							c.Identity("<integration>", "12dfb90b-e47b-4d49-b834-294b01992ad0")
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
					}),
				)
			},
		}
		return mt, Begin(mt, app).EnableHandlers("<integration>")
	}

	t.Run("func Not()", func(t *testing.T) {
		cases := []struct {
			Name        string
			Expectation Expectation
			Passes      bool
			Report      reportMatcher
		}{
			{
				"it fails when the child expectation passes",
				Not(ToRecordEvent(EventA1)),
				expectFail,
				expectReport(
					`✗ do not record a specific '*stubs.EventStub[TypeA]' event`,
				),
			},
			{
				"it passes when the child expectation fails",
				Not(ToRecordEvent(EventA2)),
				expectPass,
				expectReport(
					`✓ do not record a specific '*stubs.EventStub[TypeA]' event`,
				),
			},
		}

		for _, c := range cases {
			c := c
			t.Run(c.Name, func(t *testing.T) {
				mt, tc := newFixture()
				tc.Expect(ExecuteCommand(CommandA1), c.Expectation)
				c.Report(mt)
				if mt.Failed() != !c.Passes {
					t.Fatalf("testingT.Failed() = %v, want %v", mt.Failed(), !c.Passes)
				}
			})
		}

		t.Run("it produces the expected caption", func(t *testing.T) {
			mt, tc := newFixture()
			tc.Expect(
				noop,
				Not(ToRecordEvent(EventA2)),
			)
			xtesting.ExpectContains[string](
				t,
				"expected log message not found",
				mt.Logs,
				"--- expect [no-op] not to record a specific '*stubs.EventStub[TypeA]' event ---",
			)
		})

		t.Run("it fails the test if the child cannot construct a predicate", func(t *testing.T) {
			mt, tc := newFixture()
			tc.Expect(
				noop,
				Not(failBeforeAction),
			)
			xtesting.ExpectContains[string](
				t,
				"expected log message not found",
				mt.Logs,
				"<always fail before action>",
			)
			if !mt.Failed() {
				t.Fatal("expected test to fail")
			}
		})
	})
}
