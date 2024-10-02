package logging_test

import (
	"github.com/dogmatiq/configkit"
	. "github.com/dogmatiq/testkit/fact/internal/logging"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

var _ = g.Describe("type Icon", func() {
	g.Describe("func String()", func() {
		g.It("returns the icon string", func() {
			gm.Expect(
				TransactionIDIcon.String(),
			).To(gm.Equal("⨀"))
		})
	})
})

var _ = g.Describe("type IconWithLabel", func() {
	g.Describe("func IconWithLabel()", func() {
		g.It("returns the icon and label", func() {
			gm.Expect(
				TransactionIDIcon.WithLabel("<foo>").String(),
			).To(gm.Equal("⨀ <foo>"))
		})
	})
})

var _ = g.Describe("func DirectionIcon()", func() {
	g.It("returns the expected icon", func() {
		gm.Expect(DirectionIcon(true, false)).To(gm.Equal(InboundIcon))
		gm.Expect(DirectionIcon(false, false)).To(gm.Equal(OutboundIcon))
	})

	g.It("returns the expected error icon", func() {
		gm.Expect(DirectionIcon(true, true)).To(gm.Equal(InboundErrorIcon))
		gm.Expect(DirectionIcon(false, true)).To(gm.Equal(OutboundErrorIcon))
	})
})

var _ = g.Describe("func HandlerTypeIcon()", func() {
	g.It("returns the expected icon", func() {
		gm.Expect(HandlerTypeIcon(configkit.AggregateHandlerType)).To(gm.Equal(AggregateIcon))
		gm.Expect(HandlerTypeIcon(configkit.ProcessHandlerType)).To(gm.Equal(ProcessIcon))
		gm.Expect(HandlerTypeIcon(configkit.IntegrationHandlerType)).To(gm.Equal(IntegrationIcon))
		gm.Expect(HandlerTypeIcon(configkit.ProjectionHandlerType)).To(gm.Equal(ProjectionIcon))
	})
})
