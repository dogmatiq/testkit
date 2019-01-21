package process_test

import (
	. "github.com/dogmatiq/dogmatest/engine/controller/process"
	handlerkit "github.com/dogmatiq/dogmatest/internal/enginekit/handler"
	"github.com/dogmatiq/dogmatest/internal/fixtures"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Controller", func() {
	var (
		handler    *fixtures.ProcessMessageHandler
		controller *Controller
	)

	BeforeEach(func() {
		handler = &fixtures.ProcessMessageHandler{}
		controller = NewController("<name>", handler)
	})

	Describe("func Name()", func() {
		It("returns the handler name", func() {
			Expect(controller.Name()).To(Equal("<name>"))
		})
	})

	Describe("func Type()", func() {
		It("returns handler.ProcessType", func() {
			Expect(controller.Type()).To(Equal(handlerkit.ProcessType))
		})
	})
})
