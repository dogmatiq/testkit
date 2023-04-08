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
				c.Identity("<handler-name>", "ca76057c-9ad0-4a55-a9d9-7fbffe92e644")
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
				c.Identity("<app>", "d61d15c0-0df7-466b-b0cc-749084399d73")
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
				c.Identity("<handler-name>", "191580b7-0b16-4e5e-be03-eda07e92b9b0")
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
				c.Identity("<app>", "ad2a18d6-d87a-4b5c-a396-aa293ec64fdf")
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
