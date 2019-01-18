package render

import (
	"io"
	"testing"

	"github.com/dogmatiq/dogmatest/internal/fixtures"
	"github.com/dogmatiq/iago"
	"github.com/dogmatiq/iago/iotest"
)

func TestDefaultRenderer_WriteMessage(t *testing.T) {
	iotest.TestWrite(
		t,
		func(w io.Writer) int {
			return iago.Must(
				DefaultRenderer{}.WriteMessage(w, fixtures.Message{
					Value: "<value>",
				}),
			)
		},
		"fixtures.Message{",
		`    Value: "<value>"`,
		"}",
	)
}

func TestDefaultRenderer_WriteAggregateRoot(t *testing.T) {
	iotest.TestWrite(
		t,
		func(w io.Writer) int {
			return iago.Must(
				DefaultRenderer{}.WriteAggregateRoot(w, &fixtures.AggregateRoot{
					Value: "<value>",
				}),
			)
		},
		"*fixtures.AggregateRoot{",
		`    Value: "<value>"`,
		"}",
	)
}

func TestDefaultRenderer_WriteProcessRoot(t *testing.T) {
	iotest.TestWrite(
		t,
		func(w io.Writer) int {
			return iago.Must(
				DefaultRenderer{}.WriteProcessRoot(w, &fixtures.ProcessRoot{
					Value: "<value>",
				}),
			)
		},
		"*fixtures.ProcessRoot{",
		`    Value: "<value>"`,
		"}",
	)
}
