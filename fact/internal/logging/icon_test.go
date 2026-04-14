package logging_test

import (
	"testing"

	"github.com/dogmatiq/enginekit/config"
	. "github.com/dogmatiq/testkit/fact/internal/logging"
	"github.com/dogmatiq/testkit/internal/test"
)

func TestIcon(t *testing.T) {
	t.Run("func String()", func(t *testing.T) {
		t.Run("it returns the icon string", func(t *testing.T) {
			test.Expect(
				t,
				"unexpected icon string",
				TransactionIDIcon.String(),
				"⨀",
			)
		})
	})
}

func TestIconWithLabel(t *testing.T) {
	t.Run("it returns the icon and label", func(t *testing.T) {
		test.Expect(
			t,
			"unexpected icon label string",
			TransactionIDIcon.WithLabel("<foo>").String(),
			"⨀ <foo>",
		)
	})
}

func TestDirectionIcon(t *testing.T) {
	t.Run("it returns the expected icon", func(t *testing.T) {
		test.Expect(t, "unexpected inbound icon", DirectionIcon(true, false), InboundIcon)
		test.Expect(t, "unexpected outbound icon", DirectionIcon(false, false), OutboundIcon)
	})

	t.Run("it returns the expected error icon", func(t *testing.T) {
		test.Expect(t, "unexpected inbound error icon", DirectionIcon(true, true), InboundErrorIcon)
		test.Expect(t, "unexpected outbound error icon", DirectionIcon(false, true), OutboundErrorIcon)
	})
}

func TestHandlerTypeIcon(t *testing.T) {
	t.Run("it returns the expected icon", func(t *testing.T) {
		test.Expect(t, "unexpected aggregate icon", HandlerTypeIcon(config.AggregateHandlerType), AggregateIcon)
		test.Expect(t, "unexpected process icon", HandlerTypeIcon(config.ProcessHandlerType), ProcessIcon)
		test.Expect(t, "unexpected integration icon", HandlerTypeIcon(config.IntegrationHandlerType), IntegrationIcon)
		test.Expect(t, "unexpected projection icon", HandlerTypeIcon(config.ProjectionHandlerType), ProjectionIcon)
	})
}
