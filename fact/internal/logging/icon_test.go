package logging_test

import (
	"github.com/dogmatiq/configkit"
	. "github.com/dogmatiq/testkit/fact/internal/logging"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("type Icon", func() {
	g.Describe("func String()", func() {
		g.It("returns the icon string", func() {
			Expect(
				TransactionIDIcon.String(),
			).To(Equal("⨀"))
		})
	})
})

var _ = g.Describe("type IconWithLabel", func() {
	g.Describe("func IconWithLabel()", func() {
		g.It("returns the icon and label", func() {
			Expect(
				TransactionIDIcon.WithLabel("<foo>").String(),
			).To(Equal("⨀ <foo>"))
		})
	})
})

var _ = g.Describe("func DirectionIcon()", func() {
	g.It("returns the expected icon", func() {
		Expect(DirectionIcon(true, false)).To(Equal(InboundIcon))
		Expect(DirectionIcon(false, false)).To(Equal(OutboundIcon))
	})

	g.It("returns the expected error icon", func() {
		Expect(DirectionIcon(true, true)).To(Equal(InboundErrorIcon))
		Expect(DirectionIcon(false, true)).To(Equal(OutboundErrorIcon))
	})
})

var _ = g.Describe("func HandlerTypeIcon()", func() {
	g.It("returns the expected icon", func() {
		Expect(HandlerTypeIcon(configkit.AggregateHandlerType)).To(Equal(AggregateIcon))
		Expect(HandlerTypeIcon(configkit.ProcessHandlerType)).To(Equal(ProcessIcon))
		Expect(HandlerTypeIcon(configkit.IntegrationHandlerType)).To(Equal(IntegrationIcon))
		Expect(HandlerTypeIcon(configkit.ProjectionHandlerType)).To(Equal(ProjectionIcon))
	})
})
