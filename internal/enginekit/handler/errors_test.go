package handler_test

import (
	. "github.com/dogmatiq/dogmatest/internal/enginekit/handler"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type EmptyInstanceIDError", func() {
	Describe("func Error", func() {
		It("returns a meaningful error description", func() {
			err := EmptyInstanceIDError{
				HandlerName: "<name>",
				HandlerType: AggregateType,
			}

			Expect(err.Error()).To(Equal(
				"the '<name>' aggregate message handler attempted to route a message to an empty instance ID",
			))
		})
	})
})

var _ = Describe("type NilRootError", func() {
	Describe("func Error", func() {
		It("returns a meaningful error description", func() {
			err := NilRootError{
				HandlerName: "<name>",
				HandlerType: AggregateType,
			}

			Expect(err.Error()).To(Equal(
				"the '<name>' aggregate message handler produced a nil root",
			))
		})
	})
})
