package assert_test

import (
	"context"
	"strings"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	"github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/internal/testingmock"
	"github.com/onsi/gomega"
)

func newTestApp() (
	*Application,
	*AggregateMessageHandler,
	*ProcessMessageHandler,
	*IntegrationMessageHandler,
) {
	aggregate := &AggregateMessageHandler{
		ConfigureFunc: func(c dogma.AggregateConfigurer) {
			c.Identity("<aggregate>", "<aggregate-key>")
			c.ConsumesCommandType(MessageA{})
			c.ProducesEventType(MessageB{})
		},
		RouteCommandToInstanceFunc: func(dogma.Message) string {
			return "<aggregate-instance>"
		},
		HandleCommandFunc: func(
			_ dogma.AggregateRoot,
			s dogma.AggregateCommandScope,
			m dogma.Message,
		) {
			s.RecordEvent(
				MessageB{Value: "<value>"},
			)
		},
	}

	process := &ProcessMessageHandler{
		ConfigureFunc: func(c dogma.ProcessConfigurer) {
			c.Identity("<process>", "<process-key>")
			c.ConsumesEventType(MessageB{})
			c.ProducesCommandType(MessageC{})
		},
		RouteEventToInstanceFunc: func(context.Context, dogma.Message) (string, bool, error) {
			return "<process-instance>", true, nil
		},
		HandleEventFunc: func(
			_ context.Context,
			s dogma.ProcessEventScope,
			m dogma.Message,
		) error {
			s.Begin()
			s.ExecuteCommand(
				MessageC{Value: "<value>"},
			)

			return nil
		},
	}

	integration := &IntegrationMessageHandler{
		ConfigureFunc: func(c dogma.IntegrationConfigurer) {
			c.Identity("<integration>", "<integration-key>")
			c.ConsumesCommandType(MessageC{})
			c.ProducesEventType(MessageD{})
		},
		HandleCommandFunc: func(
			_ context.Context,
			s dogma.IntegrationCommandScope,
			m dogma.Message,
		) error {
			s.RecordEvent(
				MessageD{Value: "<value>"},
			)

			return nil
		},
	}

	app := &Application{
		ConfigureFunc: func(c dogma.ApplicationConfigurer) {
			c.Identity("<app>", "<app-key>")
			c.RegisterAggregate(aggregate)
			c.RegisterProcess(process)
			c.RegisterIntegration(integration)
		},
	}

	return app, aggregate, process, integration
}

func runTest(
	app dogma.Application,
	op func(*testkit.Test),
	options []engine.OperationOption,
	expectOk bool,
	expectReport []string,
) {
	t := &testingmock.T{
		FailSilently: true,
	}

	opts := append(
		[]engine.OperationOption{
			engine.EnableAggregates(true),
			engine.EnableProcesses(true),
			engine.EnableIntegrations(true),
			engine.EnableProjections(true),
		},
		options...,
	)

	test := testkit.Begin(
		t,
		app,
		testkit.WithUnsafeOperationOptions(opts...),
	)

	op(test)

	logs := strings.TrimSpace(strings.Join(t.Logs, "\n"))
	lines := strings.Split(logs, "\n")

	for i, l := range lines {
		if l == "--- TEST REPORT ---" {
			gomega.Expect(lines[i:]).To(gomega.Equal(expectReport))
			gomega.Expect(t.Failed()).To(gomega.Equal(!expectOk))
			return
		}
	}

	gomega.Expect(lines).To(gomega.Equal(expectReport))
	gomega.Expect(t.Failed()).To(gomega.Equal(!expectOk))
}
