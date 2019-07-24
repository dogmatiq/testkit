package render

import (
	"io"

	"github.com/dogmatiq/enginekit/fixtures"
	"github.com/dogmatiq/iago/iotest"
	"github.com/dogmatiq/iago/must"
	. "github.com/onsi/ginkgo"
)

var _ Renderer = DefaultRenderer{}

var _ = Describe("type DefaultRenderer", func() {
	Describe("func WriteMessage", func() {
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

	Describe("func WriteAggregateRoot", func() {
		It("writes a suitable representation", func() {
			iotest.TestWrite(
				GinkgoT(),
				func(w io.Writer) int {
					return must.Must(
						DefaultRenderer{}.WriteAggregateRoot(
							w,
							&fixtures.AggregateRoot{Value: "<value>"},
						),
					)
				},
				"*fixtures.AggregateRoot{",
				`    Value:          "<value>"`,
				`    ApplyEventFunc: nil`,
				"}",
			)
		})
	})

	Describe("func WriteProcessRoot", func() {
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

	Describe("func WriteAggregateMessageHandler", func() {
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
				"*fixtures.AggregateMessageHandler{",
				"    NewFunc:                    nil",
				"    ConfigureFunc:              nil",
				"    RouteCommandToInstanceFunc: nil",
				"    HandleCommandFunc:          nil",
				"}",
			)
		})
	})

	Describe("func WriteProcessMessageHandler", func() {
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
				"*fixtures.ProcessMessageHandler{",
				"    NewFunc:                  nil",
				"    ConfigureFunc:            nil",
				"    RouteEventToInstanceFunc: nil",
				"    HandleEventFunc:          nil",
				"    HandleTimeoutFunc:        nil",
				"    TimeoutHintFunc:          nil",
				"}",
			)
		})
	})

	Describe("func WriteIntegrationMessageHandler", func() {
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
				"*fixtures.IntegrationMessageHandler{",
				"    ConfigureFunc:     nil",
				"    HandleCommandFunc: nil",
				"    TimeoutHintFunc:   nil",
				"}",
			)
		})
	})

	Describe("func WriteProjectionMessageHandler", func() {
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
				"*fixtures.ProjectionMessageHandler{",
				"    ConfigureFunc:   nil",
				"    HandleEventFunc: nil",
				"    RecoverFunc:     nil",
				"    DiscardFunc:     nil",
				"    TimeoutHintFunc: nil",
				"}",
			)
		})
	})
})
