package panicx_test

import (
	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/engine/internal/panicx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type UnexpectedBehavior", func() {
	config := configkit.FromProjection(
		&ProjectionMessageHandler{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<name>", "<key>")
				c.ConsumesEventType(MessageE{})
			},
		},
	)

	Describe("func String()", func() {
		It("returns a description of the panic", func() {
			x := UnexpectedBehavior{
				Handler:        config,
				Interface:      "<interface>",
				Method:         "<method>",
				Implementation: config.Handler(),
				Description:    "<description>",
			}

			Expect(x.String()).To(Equal(
				"the '<name>' projection message handler behaved unexpectedly in *fixtures.ProjectionMessageHandler.<method>(): <description>",
			))
		})
	})
})
