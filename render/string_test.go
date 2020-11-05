package render_test

import (
	"strings"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogma/fixtures" // can't dot-import due to conflicts
	. "github.com/dogmatiq/testkit/render"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("func Message()", func() {
	It("returns a suitable representation", func() {
		Expect(
			Message(
				DefaultRenderer{},
				fixtures.MessageA1,
			),
		).To(Equal(join(
			"fixtures.MessageA{",
			`    Value: "A1"`,
			"}",
		)))
	})
})

var _ = Describe("func AggregateRoot()", func() {
	It("returns a suitable representation", func() {
		Expect(
			AggregateRoot(
				DefaultRenderer{},
				&fixtures.AggregateRoot{
					AppliedEvents: []dogma.Message{
						fixtures.MessageA1,
					},
				},
			),
		).To(Equal(join(
			"*fixtures.AggregateRoot{",
			`    AppliedEvents:  {`,
			`        fixtures.MessageA{`,
			`            Value: "A1"`,
			`        }`,
			`    }`,
			`    ApplyEventFunc: nil`,
			"}",
		)))
	})
})

var _ = Describe("func ProcessRoot()", func() {
	It("returns a suitable representation", func() {
		Expect(
			ProcessRoot(
				DefaultRenderer{},
				&fixtures.ProcessRoot{Value: "<value>"},
			),
		).To(Equal(join(
			"*fixtures.ProcessRoot{",
			`    Value: "<value>"`,
			"}",
		)))
	})
})

var _ = Describe("func AggregateMessageHandler()", func() {
	It("returns a suitable representation", func() {
		Expect(
			AggregateMessageHandler(
				DefaultRenderer{},
				&fixtures.AggregateMessageHandler{},
			),
		).To(Equal(
			"*fixtures.AggregateMessageHandler{<zero>}",
		))
	})
})

var _ = Describe("func ProcessMessageHandler()", func() {
	It("returns a suitable representation", func() {
		Expect(
			ProcessMessageHandler(
				DefaultRenderer{},
				&fixtures.ProcessMessageHandler{},
			),
		).To(Equal(
			"*fixtures.ProcessMessageHandler{<zero>}",
		))
	})
})

var _ = Describe("func IntegrationMessageHandler()", func() {
	It("returns a suitable representation", func() {
		Expect(
			IntegrationMessageHandler(
				DefaultRenderer{},
				&fixtures.IntegrationMessageHandler{},
			),
		).To(Equal(
			"*fixtures.IntegrationMessageHandler{<zero>}",
		))
	})
})

var _ = Describe("func ProjectionMessageHandler()", func() {
	It("returns a suitable representation", func() {
		Expect(
			ProjectionMessageHandler(
				DefaultRenderer{},
				&fixtures.ProjectionMessageHandler{},
			),
		).To(Equal(
			"*fixtures.ProjectionMessageHandler{<zero>}",
		))
	})
})

func join(values ...string) string {
	return strings.Join(values, "\n")
}
