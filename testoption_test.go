package testkit

import (
	"context"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine"

	"github.com/dogmatiq/enginekit/fixtures"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("func StartTime()", func() {
	It("sets the engine time as seen during a Prepare() call", func() {
		now := time.Date(2001, 2, 3, 4, 5, 6, 7, time.UTC)
		called := false

		handler := &fixtures.ProjectionMessageHandler{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<handler-name>", "<handler-key>")
				c.ConsumesEventType(fixtures.MessageA{})
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

		app := &fixtures.Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "<app-key>")
				c.RegisterProjection(handler)
			},
		}

		New(app).
			Begin(
				&mockT{},
				StartTime(now),
				WithOperationOptions(
					engine.EnableProjections(true),
				),
			).
			Prepare(fixtures.MessageA1)

		Expect(called).To(BeTrue())
	})
})
