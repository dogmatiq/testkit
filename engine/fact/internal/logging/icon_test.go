package logging_test

import (
	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/enginekit/message"
	. "github.com/dogmatiq/testkit/engine/fact/internal/logging"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Icon", func() {
	Describe("func String()", func() {
		It("returns the icon string", func() {
			Expect(
				TransactionIDIcon.String(),
			).To(Equal("⨀"))
		})
	})
})

var _ = Describe("type IconWithLabel", func() {
	Describe("func IconWithLabel()", func() {
		It("returns the icon and label", func() {
			Expect(
				TransactionIDIcon.WithLabel("<foo>").String(),
			).To(Equal("⨀ <foo>"))
		})
	})
})

var _ = Describe("func DirectionIcon()", func() {
	It("returns the expected icon", func() {
		Expect(DirectionIcon(message.InboundDirection, false)).To(Equal(InboundIcon))
		Expect(DirectionIcon(message.OutboundDirection, false)).To(Equal(OutboundIcon))
	})

	It("returns the expected error icon", func() {
		Expect(DirectionIcon(message.InboundDirection, true)).To(Equal(InboundErrorIcon))
		Expect(DirectionIcon(message.OutboundDirection, true)).To(Equal(OutboundErrorIcon))
	})
})

var _ = Describe("func HandlerTypeIcon()", func() {
	It("returns the expected icon", func() {
		Expect(HandlerTypeIcon(configkit.AggregateHandlerType)).To(Equal(AggregateIcon))
		Expect(HandlerTypeIcon(configkit.ProcessHandlerType)).To(Equal(ProcessIcon))
		Expect(HandlerTypeIcon(configkit.IntegrationHandlerType)).To(Equal(IntegrationIcon))
		Expect(HandlerTypeIcon(configkit.ProjectionHandlerType)).To(Equal(ProjectionIcon))
	})
})
