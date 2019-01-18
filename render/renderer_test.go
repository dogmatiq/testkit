package render

import (
	"io"
	"testing"

	"github.com/dogmatiq/dogmatest/internal/fixtures"
	"github.com/dogmatiq/iago"
	"github.com/dogmatiq/iago/iotest"
)

var _ Renderer = DefaultRenderer{}

func TestDefaultRenderer_WriteMessage(t *testing.T) {
	iotest.TestWrite(
		t,
		func(w io.Writer) int {
			return iago.Must(
				DefaultRenderer{}.WriteMessage(
					w,
					fixtures.MessageA{Value: "<value>"},
				),
			)
		},
		"fixtures.MessageA{",
		`    Value: "<value>"`,
		"}",
	)
}

func TestDefaultRenderer_WriteAggregateRoot(t *testing.T) {
	iotest.TestWrite(
		t,
		func(w io.Writer) int {
			return iago.Must(
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
}

func TestDefaultRenderer_WriteProcessRoot(t *testing.T) {
	iotest.TestWrite(
		t,
		func(w io.Writer) int {
			return iago.Must(
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
}

func TestDefaultRenderer_WriteAggregateMessageHandler(t *testing.T) {
	iotest.TestWrite(
		t,
		func(w io.Writer) int {
			return iago.Must(
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
}

func TestDefaultRenderer_WriteProcessMessageHandler(t *testing.T) {
	iotest.TestWrite(
		t,
		func(w io.Writer) int {
			return iago.Must(
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
		"}",
	)
}

func TestDefaultRenderer_WriteIntegrationMessageHandler(t *testing.T) {
	iotest.TestWrite(
		t,
		func(w io.Writer) int {
			return iago.Must(
				DefaultRenderer{}.WriteIntegrationMessageHandler(
					w,
					&fixtures.IntegrationMessageHandler{},
				),
			)
		},
		"*fixtures.IntegrationMessageHandler{",
		"    ConfigureFunc:     nil",
		"    HandleCommandFunc: nil",
		"}",
	)
}

func TestDefaultRenderer_WriteProjectionMessageHandler(t *testing.T) {
	iotest.TestWrite(
		t,
		func(w io.Writer) int {
			return iago.Must(
				DefaultRenderer{}.WriteProjectionMessageHandler(
					w,
					&fixtures.ProjectionMessageHandler{},
				),
			)
		},
		"*fixtures.ProjectionMessageHandler{",
		"    ConfigureFunc:   nil",
		"    HandleEventFunc: nil",
		"}",
	)
}
