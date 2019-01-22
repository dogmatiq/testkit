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

var _ = Describe("type EventNotRecordedError", func() {
	Describe("func Error", func() {
		When("the instance was created", func() {
			It("returns a meaningful error description", func() {
				err := EventNotRecordedError{
					HandlerName:  "<name>",
					InstanceID:   "<instance>",
					WasDestroyed: false,
				}

				Expect(err.Error()).To(Equal(
					"the '<name>' aggregate message handler created the '<instance>' instance without recording an event",
				))
			})
		})

		When("the instance was destroyed", func() {
			It("returns a meaningful error description", func() {
				err := EventNotRecordedError{
					HandlerName:  "<name>",
					InstanceID:   "<instance>",
					WasDestroyed: true,
				}

				Expect(err.Error()).To(Equal(
					"the '<name>' aggregate message handler destroyed the '<instance>' instance without recording an event",
				))
			})
		})
	})
})
