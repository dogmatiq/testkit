package envelope_test

import (
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
	"github.com/dogmatiq/dogmatest/internal/fixtures"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Envelope", func() {
	Describe("func New", func() {
		It("returns the expected envelope", func() {
			m := fixtures.MessageA{Value: "<value>"}
			env := New(m, message.CommandRole)

			Expect(env).To(Equal(
				&Envelope{
					Message: m,
					Type:    message.TypeOf(m),
					Role:    message.CommandRole,
					IsRoot:  true,
				},
			))
		})
	})

	Describe("func NewChild", func() {
		var (
			parent *Envelope
			m      dogma.Message
		)

		BeforeEach(func() {
			parent = New(
				fixtures.MessageA{Value: "<parent>"},
				message.CommandRole,
			)

			m = fixtures.MessageA{Value: "<parent>"}
		})

		It("returns the expected envelope", func() {
			env := parent.NewChild(m, message.EventRole)
			Expect(env).To(Equal(
				&Envelope{
					Message: m,
					Type:    message.TypeOf(m),
					Role:    message.EventRole,
					IsRoot:  false,
				},
			))

		})

		It("adds the child envelope to the parent", func() {
			env := parent.NewChild(m, message.EventRole)
			Expect(parent.Children).To(Equal(
				[]*Envelope{env},
			))
		})
	})

	Describe("func Walk", func() {
		var root *Envelope

		BeforeEach(func() {
			root = New(
				fixtures.MessageA{Value: "<root>"},
				message.CommandRole,
			)

			child1 := root.NewChild(
				fixtures.MessageA{Value: "<child-1>"},
				message.CommandRole,
			)

			child1.NewChild(
				fixtures.MessageA{Value: "<grandchild-1a>"},
				message.CommandRole,
			)

			child1.NewChild(
				fixtures.MessageA{Value: "<grandchild-1b>"},
				message.CommandRole,
			)

			child2 := root.NewChild(
				fixtures.MessageA{Value: "<child-1>"},
				message.CommandRole,
			)

			child2.NewChild(
				fixtures.MessageA{Value: "<grandchild-2a>"},
				message.CommandRole,
			)

			child2.NewChild(
				fixtures.MessageA{Value: "<grandchild-2a>"},
				message.CommandRole,
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
