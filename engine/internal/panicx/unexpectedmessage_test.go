package panicx_test

import (
	"strings"
	"testing"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config/runtimeconfig"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/engine/internal/panicx"
	"github.com/dogmatiq/testkit/internal/x/xtesting"
)

func TestUnexpectedMessage(t *testing.T) {
	config := runtimeconfig.FromProjection(
		&ProjectionMessageHandlerStub{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<name>", "a0eab8dd-db22-467a-87c2-c38138c582e8")
				c.Routes(
					dogma.HandlesEvent[*EventStub[TypeA]](),
				)
			},
		},
	)

	t.Run("func String()", func(t *testing.T) {
		defer func() {
			r := recover()
			x, ok := r.(UnexpectedMessage)
			if !ok {
				t.Fatalf("expected UnexpectedMessage panic, got %T", r)
			}

			xtesting.Expect(
				t,
				"unexpected string representation",
				x.String(),
				"the '<name>' projection message handler did not expect *stubs.ProjectionMessageHandlerStub.<method>() to be called with a message of type *stubs.EventStub[TypeX]",
			)
		}()

		EnrichUnexpectedMessage(
			config,
			"<interface>",
			"<method>",
			config.Source.Get(),
			EventX1,
			func() {
				panic(dogma.UnexpectedMessage)
			},
		)
	})
}

func TestEnrichUnexpectedMessage(t *testing.T) {
	config := runtimeconfig.FromProjection(
		&ProjectionMessageHandlerStub{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<name>", "b665eca3-936e-41e3-b9ab-c618cfa95ec2")
				c.Routes(
					dogma.HandlesEvent[*EventStub[TypeA]](),
				)
			},
		},
	)

	t.Run("calls the function", func(t *testing.T) {
		called := false

		EnrichUnexpectedMessage(
			config,
			"<interface>",
			"<method>",
			config.Source.Get(),
			EventX1,
			func() {
				called = true
			},
		)

		xtesting.Expect(t, "expected function to be called", called, true)
	})

	t.Run("propagates panic values", func(t *testing.T) {
		defer func() {
			r := recover()
			xtesting.Expect(t, "unexpected panic value", r, "<panic>")
		}()

		EnrichUnexpectedMessage(
			config,
			"<interface>",
			"<method>",
			config.Source.Get(),
			EventX1,
			func() {
				panic("<panic>")
			},
		)
	})

	t.Run("converts UnexpectedMessage values", func(t *testing.T) {
		defer func() {
			r := recover()
			x, ok := r.(UnexpectedMessage)
			if !ok {
				t.Fatalf("expected UnexpectedMessage panic, got %T", r)
			}

			xtesting.Expect(t, "unexpected handler", x.Handler, config)
			xtesting.Expect(t, "unexpected interface", x.Interface, "<interface>")
			xtesting.Expect(t, "unexpected method", x.Method, "<method>")
			xtesting.Expect(t, "unexpected implementation", x.Implementation, config.Source.Get())
			xtesting.Expect(t, "unexpected message", x.Message, EventX1)
			xtesting.Expect(
				t,
				"unexpected panic location func",
				x.PanicLocation.Func,
				"github.com/dogmatiq/testkit/engine/internal/panicx_test.doPanic",
			)
			if !strings.HasSuffix(x.PanicLocation.File, "/engine/internal/panicx/linenumber_test.go") {
				t.Fatalf("unexpected panic location file: %s", x.PanicLocation.File)
			}
			xtesting.Expect(t, "unexpected panic location line", x.PanicLocation.Line, 50)
		}()

		EnrichUnexpectedMessage(
			config,
			"<interface>",
			"<method>",
			config.Source.Get(),
			EventX1,
			doPanic,
		)
	})
}
