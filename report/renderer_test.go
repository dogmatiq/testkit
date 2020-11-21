package report

import (
	"io"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogma/fixtures" // can't dot-import due to conflicts
	"github.com/dogmatiq/iago/iotest"
	"github.com/dogmatiq/iago/must"
	. "github.com/onsi/ginkgo"
)

var _ Renderer = DefaultRenderer{}

var _ = Describe("type DefaultRenderer", func() {
	Describe("func WriteMessage()", func() {
		It("writes a suitable representation", func() {
			iotest.TestWrite(
				GinkgoT(),
				func(w io.Writer) int {
					return must.Must(
						DefaultRenderer{}.WriteMessage(
							w,
							fixtures.MessageA1,
						),
					)
				},
				"fixtures.MessageA{",
				`    Value: "A1"`,
				"}",
			)
		})
	})

	Describe("func WriteAggregateRoot()", func() {
		It("writes a suitable representation", func() {
			iotest.TestWrite(
				GinkgoT(),
				func(w io.Writer) int {
					return must.Must(
						DefaultRenderer{}.WriteAggregateRoot(
							w,
							&fixtures.AggregateRoot{
								AppliedEvents: []dogma.Message{
									fixtures.MessageA1,
								},
							},
						),
					)
				},
				"*fixtures.AggregateRoot{",
				`    AppliedEvents:  {`,
				`        fixtures.MessageA{`,
				`            Value: "A1"`,
				`        }`,
				`    }`,
				`    ApplyEventFunc: nil`,
				"}",
			)
		})
	})

	Describe("func WriteProcessRoot()", func() {
		It("writes a suitable representation", func() {
			iotest.TestWrite(
				GinkgoT(),
				func(w io.Writer) int {
					return must.Must(
						DefaultRenderer{}.WriteProcessRoot(
							w,
							&fixtures.ProcessRoot{Value: "<value>"},
						),
					)
				},
				"*fixtures.ProcessRoot{",
				`    Value: "<value>"`,
				"}",
			)
		})
	})

	Describe("func WriteAggregateMessageHandler()", func() {
		It("writes a suitable representation", func() {
			iotest.TestWrite(
				GinkgoT(),
				func(w io.Writer) int {
					return must.Must(
						DefaultRenderer{}.WriteAggregateMessageHandler(
							w,
							&fixtures.AggregateMessageHandler{},
						),
					)
				},
				"*fixtures.AggregateMessageHandler{<zero>}",
			)
		})
	})

	Describe("func WriteProcessMessageHandler()", func() {
		It("writes a suitable representation", func() {
			iotest.TestWrite(
				GinkgoT(),
				func(w io.Writer) int {
					return must.Must(
						DefaultRenderer{}.WriteProcessMessageHandler(
							w,
							&fixtures.ProcessMessageHandler{},
						),
					)
				},
				"*fixtures.ProcessMessageHandler{<zero>}",
			)
		})
	})

	Describe("func WriteIntegrationMessageHandler()", func() {
		It("writes a suitable representation", func() {
			iotest.TestWrite(
				GinkgoT(),
				func(w io.Writer) int {
					return must.Must(
						DefaultRenderer{}.WriteIntegrationMessageHandler(
							w,
							&fixtures.IntegrationMessageHandler{},
						),
					)
				},
				"*fixtures.IntegrationMessageHandler{<zero>}",
			)
		})
	})

	Describe("func WriteProjectionMessageHandler()", func() {
		It("writes a suitable representation", func() {
			iotest.TestWrite(
				GinkgoT(),
				func(w io.Writer) int {
					return must.Must(
						DefaultRenderer{}.WriteProjectionMessageHandler(
							w,
							&fixtures.ProjectionMessageHandler{},
						),
					)
				},
				"*fixtures.ProjectionMessageHandler{<zero>}",
			)
		})
	})
})
