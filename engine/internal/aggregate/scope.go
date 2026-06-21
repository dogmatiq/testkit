package aggregate

import (
	"fmt"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/message"
	"github.com/dogmatiq/testkit/engine/internal/panicx"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/compare"
	"github.com/dogmatiq/testkit/internal/validation"
	"github.com/dogmatiq/testkit/location"
)

// scope is an implementation of dogma.AggregateCommandScope.
type scope struct {
	instanceID       string
	config           *config.Aggregate
	messageIDs       *envelope.MessageIDGenerator
	observer         fact.Observer
	root, shadowRoot dogma.AggregateRoot
	lastOp           string
	events           []*envelope.Envelope
	now              time.Time
	command          *envelope.Envelope
	streamID         string
	offset           uint64
}

func (s *scope) InstanceID() string {
	s.guardAgainstDirectMutation("InstanceID", location.OfCall())
	return s.instanceID
}

func (s *scope) RecordEvent(m dogma.Event) {
	s.guardAgainstDirectMutation("RecordEvent", location.OfCall())

	mt := message.TypeOf(m)

	if !s.config.RouteSet().DirectionOf(mt).Has(config.OutboundDirection) {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "AggregateMessageHandler",
			Method:         "HandleCommand",
			Implementation: s.config.Implementation(),
			Message:        s.command.Message,
			Description:    fmt.Sprintf("recorded an event of type %s, which is not produced by this handler", mt),
			Location:       location.OfCall(),
		})
	}

	if err := m.Validate(validation.EventValidationScope()); err != nil {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "AggregateMessageHandler",
			Method:         "HandleCommand",
			Implementation: s.config.Implementation(),
			Message:        s.command.Message,
			Description:    fmt.Sprintf("recorded an invalid %s event: %s", mt, err),
			Location:       location.OfCall(),
		})
	}

	if s.offset == 0 {
		s.observer.Notify(fact.AggregateInstanceCreated{
			Handler:    s.config,
			InstanceID: s.instanceID,
			Root:       s.root,
			Envelope:   s.command,
		})
	}

	panicx.EnrichUnexpectedMessage(
		s.config,
		"AggregateRoot",
		"ApplyEvent",
		s.root,
		m,
		func() {
			s.root.ApplyEvent(m)
		},
	)

	panicx.EnrichUnexpectedMessage(
		s.config,
		"AggregateRoot",
		"ApplyEvent",
		s.shadowRoot,
		m,
		func() {
			s.shadowRoot.ApplyEvent(m)
		},
	)

	if !compare.Equal(s.root, s.shadowRoot) {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "AggregateRoot",
			Method:         "ApplyEvent",
			Implementation: s.root,
			Message:        s.command.Message,
			Description:    "non-deterministic implementation of ApplyEvent detected",
			Location:       location.OfMethod(s.root, "ApplyEvent"),
		})
	}

	env := s.command.NewEvent(
		s.messageIDs.Next(),
		m,
		s.now,
		envelope.Origin{
			Handler:     s.config,
			HandlerType: config.AggregateHandlerType,
			InstanceID:  s.instanceID,
		},
		s.streamID,
		s.offset,
	)

	s.events = append(s.events, env)
	s.offset++

	s.observer.Notify(fact.EventRecordedByAggregate{
		Handler:       s.config,
		InstanceID:    s.instanceID,
		Root:          s.root,
		Envelope:      s.command,
		EventEnvelope: env,
	})
}

func (s *scope) Now() time.Time {
	s.guardAgainstDirectMutation("Now", location.OfCall())
	return s.now
}

func (s *scope) Log(f string, v ...any) {
	s.guardAgainstDirectMutation("Log", location.OfCall())

	s.observer.Notify(fact.MessageLoggedByAggregate{
		Handler:      s.config,
		InstanceID:   s.instanceID,
		Root:         s.root,
		Envelope:     s.command,
		LogFormat:    f,
		LogArguments: v,
	})
}

// guardAgainstDirectMutation panics if the aggregate root has been modified
// directly (without using RecordEvent), then records the current scope method
// call for use in diagnostic messages.
func (s *scope) guardAgainstDirectMutation(method string, loc location.Location) {
	thisOp := ""
	if method != "" {
		thisOp = fmt.Sprintf("call to %s() at %s", method, loc)
	}

	if !compare.Equal(s.root, s.shadowRoot) {
		desc := "modified the aggregate root without using RecordEvent()"

		switch {
		case s.lastOp != "" && thisOp != "":
			desc += ", between " + s.lastOp + " and " + thisOp
		case s.lastOp != "":
			desc += ", after " + s.lastOp
		case thisOp != "":
			desc += ", before " + thisOp
		}

		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "AggregateMessageHandler",
			Method:         "HandleCommand",
			Implementation: s.config.Implementation(),
			Message:        s.command.Message,
			Description:    desc,
			Location:       location.OfCall(),
		})
	}

	s.lastOp = thisOp
}
