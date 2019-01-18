package engine_test

import (
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/engine"
	"github.com/dogmatiq/dogmatest/internal/fixtures"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Envelope", func() {
	Describe("func NewEnvelope", func() {
		It("returns the expected envelope", func() {
			m := fixtures.MessageA{Value: "<value>"}
			env := NewEnvelope(m, CommandRole)

			Expect(env).To(Equal(
				&Envelope{
					Message: m,
					Role:    CommandRole,
				},
			))
		})
	})

	Describe("func NewChild", func() {
		var (
			parent  *Envelope
			message dogma.Message
		)

		BeforeEach(func() {
			parent = NewEnvelope(
				fixtures.MessageA{Value: "<parent>"},
				CommandRole,
			)

			message = fixtures.MessageA{Value: "<parent>"}
		})

		It("returns the expected envelope", func() {
			env := parent.NewChild(message, EventRole)
			Expect(env).To(Equal(
				&Envelope{
					Message: message,
					Role:    EventRole,
				},
			))

		})

		It("adds the child envelope to the parent", func() {
			env := parent.NewChild(message, EventRole)
			Expect(parent.Children).To(Equal(
				[]*Envelope{env},
			))
		})
	})

	Describe("func Walk", func() {
		var root *Envelope

		BeforeEach(func() {
			root = NewEnvelope(
				fixtures.MessageA{Value: "<root>"},
				CommandRole,
			)

			child1 := root.NewChild(
				fixtures.MessageA{Value: "<child-1>"},
				CommandRole,
			)

			child1.NewChild(
				fixtures.MessageA{Value: "<grandchild-1a>"},
				CommandRole,
			)

			child1.NewChild(
				fixtures.MessageA{Value: "<grandchild-1b>"},
				CommandRole,
			)

			child2 := root.NewChild(
				fixtures.MessageA{Value: "<child-1>"},
				CommandRole,
			)

			child2.NewChild(
				fixtures.MessageA{Value: "<grandchild-2a>"},
				CommandRole,
			)

			child2.NewChild(
				fixtures.MessageA{Value: "<grandchild-2a>"},
				CommandRole,
			)
		})

		It("walks the entire tree, depth first", func() {
			var values []string

			root.Walk(func(env *Envelope) bool {
				v := env.Message.(fixtures.MessageA).Value.(string)
				values = append(values, v)
				return true
			})

			Expect(values).To(Equal(
				[]string{
					"<child-1>",
					"<grandchild-1a>",
					"<grandchild-1b>",
					"<child-1>",
					"<grandchild-2a>",
					"<grandchild-2a>",
				},
			))
		})

		It("aborts traversal if fn() returns false", func() {
			var values []string

			root.Walk(func(env *Envelope) bool {
				v := env.Message.(fixtures.MessageA).Value.(string)
				values = append(values, v)

				if v == "<grandchild-1b>" {
					return false
				}

				return true
			})

			Expect(values).To(Equal(
				[]string{
					"<child-1>",
					"<grandchild-1a>",
					"<grandchild-1b>",
				},
			))
		})
	})
})
