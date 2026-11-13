package envelope

import (
	"time"
	"uuid"

	"github.com/dogmatiq/dogma"
)

// Envelope is a container for a message that is handled by the test engine.
type Envelope struct {
	// MessageID is the [RFC 9562] UUID that identifies the message.
	//
	// [RFC 9562]: https://www.rfc-editor.org/rfc/rfc9562.html
	MessageID uuid.UUID

	// CausationID is the ID of the message that was being handled when the
	// message identified by MessageID was produced.
	CausationID uuid.UUID

	// CorrelationID is the ID of the "root" message that entered the
	// application to cause the message identified by MessageID, either directly
	// or indirectly.
	CorrelationID uuid.UUID

	// Message is the application-defined message that the envelope represents.
	Message dogma.Message

	// CreatedAt is the time at which the message was created.
	CreatedAt time.Time

	// ScheduledFor holds the time for which a deadline message is scheduled.
	// Its value is undefined for commands and events.
	ScheduledFor time.Time

	// Origin describes the message handler that produced this message.
	// It is nil if the message was not produced by a handler.
	Origin *Origin

	// EventStreamID is the [RFC 9562] UUID that identifies the event stream
	// to which this message belongs, if it is a [dogma.Event].
	//
	// [RFC 9562]: https://www.rfc-editor.org/rfc/rfc9562.html
	EventStreamID uuid.UUID

	// EventStreamOffset is the offset of the message within its event stream.
	EventStreamOffset uint64
}

// NewCommand constructs a new envelope containing the given command message.
//
// t is the time at which the message was created.
func NewCommand(
	id uuid.UUID,
	m dogma.Command,
	t time.Time,
) *Envelope {
	if id == uuid.Nil() {
		panic("message ID must not be the nil UUID")
	}

	return &Envelope{
		MessageID:     id,
		CausationID:   id,
		CorrelationID: id,
		Message:       m,
		CreatedAt:     t,
	}
}

// NewEvent constructs a new envelope containing the given event message.
//
// t is the time at which the message was created. streamID is the ID of the
// event stream to which the event belongs. The event is the first event in
// that stream, at offset 0.
func NewEvent(
	id uuid.UUID,
	m dogma.Event,
	t time.Time,
	streamID uuid.UUID,
) *Envelope {
	if id == uuid.Nil() {
		panic("message ID must not be the nil UUID")
	}

	return &Envelope{
		MessageID:     id,
		CausationID:   id,
		CorrelationID: id,
		Message:       m,
		CreatedAt:     t,
		EventStreamID: streamID,
	}
}

// NewCommand constructs a new envelope as a child of e, indicating that the
// command message m is caused by e.Message.
//
// t is the time at which the message was created.
func (e *Envelope) NewCommand(
	id uuid.UUID,
	m dogma.Command,
	t time.Time,
	o Origin,
) *Envelope {
	return &Envelope{
		MessageID:     id,
		CausationID:   e.MessageID,
		CorrelationID: e.CorrelationID,
		Message:       m,
		CreatedAt:     t,
		Origin:        &o,
	}
}

// NewEvent constructs a new envelope as a child of e, indicating that the event
// message m is caused by e.Message.
//
// t is the time at which the message was created.
func (e *Envelope) NewEvent(
	id uuid.UUID,
	m dogma.Event,
	t time.Time,
	o Origin,
	streamID uuid.UUID,
	offset uint64,
) *Envelope {
	return &Envelope{
		MessageID:         id,
		CausationID:       e.MessageID,
		CorrelationID:     e.CorrelationID,
		Message:           m,
		CreatedAt:         t,
		Origin:            &o,
		EventStreamID:     streamID,
		EventStreamOffset: offset,
	}
}

// NewDeadline constructs a new envelope as a child of e, indicating that the
// deadline message m is caused by e.Message.
//
// t is the time at which the message was created. s is the time for which the
// deadline is scheduled.
func (e *Envelope) NewDeadline(
	id uuid.UUID,
	m dogma.Deadline,
	t time.Time,
	s time.Time,
	o Origin,
) *Envelope {
	return &Envelope{
		MessageID:     id,
		CausationID:   e.MessageID,
		CorrelationID: e.CorrelationID,
		Message:       m,
		CreatedAt:     t,
		ScheduledFor:  s,
		Origin:        &o,
	}
}
