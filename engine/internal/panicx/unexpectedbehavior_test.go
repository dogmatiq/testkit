package panicx_test

import (
	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/engine/internal/panicx"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("type UnexpectedBehavior", func() {
	config := configkit.FromProjection(
		&ProjectionMessageHandlerStub{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<name>", "fce4f9f3-e8ee-45ce-924f-be8c3c0a9285")
				c.Routes(
					dogma.HandlesEvent[MessageE](),
				)
			},
		},
	)

	g.Describe("func String()", func() {
		g.It("returns a description of the panic", func() {
			x := UnexpectedBehavior{
				Handler:        config,
				Interface:      "<interface>",
				Method:         "<method>",
				Implementation: config.Handler(),
				Description:    "<description>",
			}

			Expect(x.String()).To(Equal(
				"the '<name>' projection message handler behaved unexpectedly in *stubs.ProjectionMessageHandlerStub.<method>(): <description>",
			))
		})
	})
})
