package panicx_test

import (
	"testing"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config/runtimeconfig"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/engine/internal/panicx"
	"github.com/dogmatiq/testkit/x/xtesting"
)

func TestUnexpectedBehavior(t *testing.T) {
	config := runtimeconfig.FromProjection(
		&ProjectionMessageHandlerStub{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<name>", "fce4f9f3-e8ee-45ce-924f-be8c3c0a9285")
				c.Routes(
					dogma.HandlesEvent[*EventStub[TypeA]](),
				)
			},
		},
	)

	t.Run("func String()", func(t *testing.T) {
		x := UnexpectedBehavior{
			Handler:        config,
			Interface:      "<interface>",
			Method:         "<method>",
			Implementation: config.Source.Get(),
			Description:    "<description>",
		}

		xtesting.Expect(
			t,
			"unexpected string representation",
			x.String(),
			"the '<name>' projection message handler behaved unexpectedly in *stubs.ProjectionMessageHandlerStub.<method>(): <description>",
		)
	})
}
