package logging

import (
	"fmt"
	"io"

	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/iago/must"
)

const (
	// TransactionIDIcon is the icon shown directly before a transaction ID. It
	// is a circle with a dot in the center, intended to be reminiscent of an
	// electron circling a nucleus, indicating "atomicity". There is a unicode
	// atom symbol, however it does not tend to be discernable at smaller font
	// sizes.
	TransactionIDIcon Icon = "⨀"

	// MessageIDIcon is the icon shown directly before a message ID. It is an
	// "equals sign", indicating that this message "has exactly" the displayed
	// ID.
	MessageIDIcon Icon = "="

	// CausationIDIcon is the icon shown directly before a message causation ID.
	// It is the mathematical "because" symbol, indicating that this message
	// happened "because of" the displayed ID.
	CausationIDIcon Icon = "∵"

	// CorrelationIDIcon is the icon shown directly before a message correlation
	// ID. It is the mathematical "member of set" symbol, indicating that this
	// message belongs to the set of messages that came about because of the
	// displayed ID.
	CorrelationIDIcon Icon = "⋲"

	// InboundIcon is the icon shown to indicate that a message is "inbound" to
	// a handler. It is a downward pointing arrow, as inbound messages could be
	// considered as being "downloaded" from the network or queue.
	InboundIcon Icon = "▼"

	// InboundErrorIcon is a variant of InboundIcon used when there is an error
	// condition. It is an hollow version of the regular inbound icon,
	// indicating that the requirement remains "unfulfilled".
	InboundErrorIcon Icon = "▽"

	// OutboundIcon is the icon shown to indicate that a message is "outbound"
	// from a handler. It is an upward pointing arrow, as outbound messages
	// could be considered as being "uploaded" to the network or queue.
	OutboundIcon Icon = "▲"

	// OutboundErrorIcon is a variant of OutboundIcon used when there is an
	// error condition. It is an hollow version of the regular inbound icon,
	// indicating that the requirement remains "unfulfilled".
	OutboundErrorIcon Icon = "△"

	// RetryIcon is an icon used instead of InboundIcon when a message is being
	// re-attempted. It is an open-circle with an arrow, indicating that the
	// message has "come around again".
	RetryIcon Icon = "↻"

	// ErrorIcon is the icon shown when logging information about an error.
	// It is a heavy cross, indicating a failure.
	ErrorIcon Icon = "✖"

	// AggregateIcon is the icon shown when a log message relates to an
	// aggregate message handler. It is the mathematical "therefore" symbol,
	// representing the decision making as a result of the message.
	AggregateIcon Icon = "∴"

	// ProcessIcon is the icon shown when a log message relates to a process
	// message handler. It is three horizontal lines, representing the step in a
	// process.
	ProcessIcon Icon = "≡"

	// IntegrationIcon is the icon shown when a log message relates to an
	// integration message handler. It is the relational algebra "join" symbol,
	// representing the integration of two systems.
	IntegrationIcon Icon = "⨝"

	// ProjectionIcon is the icon shown when a log message relates to a
	// projection message handler. It is the mathematical "sum" symbol ,
	// representing the aggregation of events.
	ProjectionIcon Icon = "Σ"

	// SystemIcon is an icon shown when a log message relates to the internals
	// of the engine. It is a sprocket, representing the inner workings of the
	// machine.
	SystemIcon Icon = "⚙"

	// SeparatorIcon is an icon used to separate strings of unrelated text
	// inside a log message. It is a large bullet, intended to have a large
	// visual impact.
	SeparatorIcon Icon = "●"
)

// Icon is a unicode symbol used as an icon in log messages.
type Icon string

func (i Icon) String() string {
	return string(i)
}

// WriteTo writes a string representation of the icon to w.
// If i is the zero-value, a single space is rendered.
func (i Icon) WriteTo(w io.Writer) (int64, error) {
	s := i.String()
	if i == "" {
		s = " "
	}

	n, err := io.WriteString(w, s)
	return int64(n), err
}

// WithLabel return an IconWithLabel containing this icon and the given label.
func (i Icon) WithLabel(f string, v ...any) IconWithLabel {
	return IconWithLabel{
		i,
		formatLabel(fmt.Sprintf(f, v...)),
	}
}

// IconWithLabel is a container for an icon and its associated text label.
type IconWithLabel struct {
	Icon  Icon
	Label string
}

func (i IconWithLabel) String() string {
	return i.Icon.String() + " " + i.Label
}

// WriteTo writes a string representation of the icon and its label to w.
func (i IconWithLabel) WriteTo(w io.Writer) (_ int64, err error) {
	defer must.Recover(&err)

	n := must.WriteTo(w, i.Icon)
	n += must.Write(w, space1)
	n += must.WriteString(w, i.Label)

	return int64(n), err
}

// formatLabel formats a label for display.
func formatLabel(label string) string {
	if label == "" {
		return "-"
	}

	return label
}

// DirectionIcon returns the icon to use for the given message direction.
func DirectionIcon(inbound bool, isError bool) Icon {
	if inbound {
		if isError {
			return InboundErrorIcon
		}

		return InboundIcon
	}

	if isError {
		return OutboundErrorIcon
	}

	return OutboundIcon
}

// HandlerTypeIcon returns the icon to use for the given handler type.
func HandlerTypeIcon(t config.HandlerType) Icon {
	return config.MapByHandlerType(
		t,
		AggregateIcon,
		ProcessIcon,
		IntegrationIcon,
		ProjectionIcon,
	)
}
