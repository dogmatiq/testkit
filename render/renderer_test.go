package render

import (
	"io"
	"testing"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/iago"
	"github.com/dogmatiq/iago/iotest"
)

type testMessage struct {
	dogma.Message
	Value string
}

type testAggregateRoot struct {
	dogma.AggregateRoot
	Value string
}

type testProcessRoot struct {
	dogma.ProcessRoot
	Value string
}

func TestDefaultRenderer_WriteMessage(t *testing.T) {
	iotest.TestWrite(
		t,
		func(w io.Writer) int {
			return iago.Must(
				DefaultRenderer{}.WriteMessage(w, testMessage{
					Value: "<value>",
				}),
			)
		},
		"render.testMessage{",
		"    Message: nil",
		`    Value:   "<value>"`,
		"}",
	)
}

func TestDefaultRenderer_WriteAggregateRoot(t *testing.T) {
	iotest.TestWrite(
		t,
		func(w io.Writer) int {
			return iago.Must(
				DefaultRenderer{}.WriteAggregateRoot(w, testAggregateRoot{
					Value: "<value>",
				}),
			)
		},
		"render.testAggregateRoot{",
		"    AggregateRoot: nil",
		`    Value:         "<value>"`,
		"}",
	)
}

func TestDefaultRenderer_WriteProcessRoot(t *testing.T) {
	iotest.TestWrite(
		t,
		func(w io.Writer) int {
			return iago.Must(
				DefaultRenderer{}.WriteProcessRoot(w, testProcessRoot{
					Value: "<value>",
				}),
			)
		},
		"render.testProcessRoot{",
		"    ProcessRoot: nil",
		`    Value:       "<value>"`,
		"}",
	)
}
