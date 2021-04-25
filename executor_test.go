package testkit_test

import (
	"context"
	"errors"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("func InterceptCommandExecutor()", func() {
	var app dogma.Application

	BeforeEach(func() {
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

		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "<app-key>")
				c.RegisterIntegration(handler)
			},
		}
	})

	When("used as a TestOption", func() {
		It("intercepts calls to ExecuteCommand()", func() {
			test := Begin(
				&testingmock.T{},
				app,
				InterceptCommandExecutor(
					func(
						ctx context.Context,
						m dogma.Message,
						e dogma.CommandExecutor,
					) error {
						Expect(m).To(Equal(MessageC1))

						err := e.ExecuteCommand(ctx, m)
						Expect(err).ShouldNot(HaveOccurred())

						return errors.New("<error>")
					},
				),
			)

			test.
				EnableHandlers("<handler-name>").
				Expect(
					Call(func() {
						err := test.CommandExecutor().ExecuteCommand(
							context.Background(),
							MessageC1,
						)
						Expect(err).To(MatchError("<error>"))
					}),
					ToRecordEvent(MessageE1),
				)
		})
	})
})
