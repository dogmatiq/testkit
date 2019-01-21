package integration_test

import (
	. "github.com/dogmatiq/dogmatest/engine/controller/integration"
	handlerkit "github.com/dogmatiq/dogmatest/internal/enginekit/handler"
	"github.com/dogmatiq/dogmatest/internal/fixtures"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Controller", func() {
	var (
		handler    *fixtures.IntegrationMessageHandler
		controller *Controller
	)

	BeforeEach(func() {
		handler = &fixtures.IntegrationMessageHandler{}
		controller = NewController("<name>", handler)
	})

	Describe("func Name()", func() {
		It("returns the handler name", func() {
			Expect(controller.Name()).To(Equal("<name>"))
		})
	})

	Describe("func Type()", func() {
		It("returns handler.IntegrationType", func() {
			Expect(controller.Type()).To(Equal(handlerkit.IntegrationType))
		})
	})
})
