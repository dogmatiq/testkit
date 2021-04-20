package testkit_test

import (
	"context"
	"time"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("func StartTimeAt()", func() {
	It("sets the engine time as seen during a Prepare() call", func() {
		now := time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC)
		called := false

		handler := &ProjectionMessageHandler{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<handler-name>", "<handler-key>")
				c.ConsumesEventType(MessageA{})
			},
			HandleEventFunc: func(
				_ context.Context,
				_, _, _ []byte,
				s dogma.ProjectionEventScope,
				_ dogma.Message,
			) (bool, error) {
				Expect(s.RecordedAt()).To(BeTemporally("==", now))
				called = true
				return true, nil
			},
		}

		app := &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "<app-key>")
				c.RegisterProjection(handler)
			},
		}

		Begin(
			&testingmock.T{},
			app,
			StartTimeAt(now),
		).
			EnableHandlers("<handler-name>").
			Prepare(RecordEvent(MessageA1))

		Expect(called).To(BeTrue())
	})
})

var _ = Describe("func WithMessageComparator()", func() {
	It("configures how messages are compared", func() {
		handler := &IntegrationMessageHandler{
			ConfigureFunc: func(c dogma.IntegrationConfigurer) {
				c.Identity("<handler-name>", "<handler-key>")
				c.ConsumesCommandType(MessageC{})
				c.ProducesEventType(MessageE{})
			},
			HandleCommandFunc: func(
				_ context.Context,
				s dogma.IntegrationCommandScope,
				_ dogma.Message,
			) error {
				s.RecordEvent(MessageE1)
				return nil
			},
		}

		app := &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "<app-key>")
				c.RegisterIntegration(handler)
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
				ExecuteCommand(MessageC1),
				ToRecordEvent(MessageE2), // this would fail without our custom comparator
			)
	})
})
