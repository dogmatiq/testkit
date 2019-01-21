package aggregate_test

import (
	. "github.com/dogmatiq/dogmatest/engine/controller/aggregate"
	handlerkit "github.com/dogmatiq/dogmatest/internal/enginekit/handler"
	"github.com/dogmatiq/dogmatest/internal/fixtures"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Controller", func() {
	var (
		handler    *fixtures.AggregateMessageHandler
		controller *Controller
	)

	BeforeEach(func() {
		handler = &fixtures.AggregateMessageHandler{}
		controller = NewController("<name>", handler)
	})

	Describe("func Name()", func() {
		It("returns the handler name", func() {
			Expect(controller.Name()).To(Equal("<name>"))
		})
	})

	Describe("func Type()", func() {
		It("returns handler.AggregateType", func() {
			Expect(controller.Type()).To(Equal(handlerkit.AggregateType))
		})
	})
})
