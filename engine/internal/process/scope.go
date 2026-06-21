package process

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

// scope is an implementation of dogma.ProcessEventScope and
// dogma.ProcessDeadlineScope.
type scope struct {
	instanceID       string
	instance         *instance
	root, shadowRoot dogma.ProcessRoot
	mutated          bool
	lastOp           string
	config           *config.Process
	handleMethod     string
	messageIDs       *envelope.MessageIDGenerator
	observer         fact.Observer
	now              time.Time
	env              *envelope.Envelope // event or deadline
	commands         []*envelope.Envelope
	ready            []*envelope.Envelope // deadlines <= now
	pending          []*envelope.Envelope // deadlines > now
}

func (s *scope) InstanceID() string {
	s.guardAgainstDirectMutation("InstanceID", location.OfCall())
	return s.instanceID
}

func (s *scope) End() {
	s.guardAgainstDirectMutation("End", location.OfCall())

	if s.instance.ended {
		return
	}

	s.instance.ended = true

	s.observer.Notify(fact.ProcessInstanceEnded{
		Handler:    s.config,
		InstanceID: s.instanceID,
		Root:       s.root,
		Envelope:   s.env,
	})
}

func (s *scope) Mutate(fn func(r dogma.ProcessRoot)) {
	s.guardAgainstDirectMutation("Mutate", location.OfCall())

	if s.instance.ended {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "ProcessMessageHandler",
			Method:         s.handleMethod,
			Implementation: s.config.Implementation(),
			Message:        s.env.Message,
			Description:    "mutated an ended process instance",
			Location:       location.OfCall(),
		})
	}

	s.mutated = true
	fn(s.root)
	fn(s.shadowRoot)

	if !compare.Equal(s.root, s.shadowRoot) {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "ProcessMessageHandler",
			Method:         s.handleMethod,
			Implementation: s.config.Implementation(),
			Message:        s.env.Message,
			Description:    "non-deterministic implementation of Mutate() callback detected",
			Location:       location.OfFunc(fn),
		})
	}
}

func (s *scope) ExecuteCommand(m dogma.Command) {
	s.guardAgainstDirectMutation("ExecuteCommand", location.OfCall())

	mt := message.TypeOf(m)

	if !s.config.RouteSet().DirectionOf(mt).Has(config.OutboundDirection) {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "ProcessMessageHandler",
			Method:         s.handleMethod,
			Implementation: s.config.Implementation(),
			Message:        s.env.Message,
			Description:    fmt.Sprintf("executed a command of type %s, which is not produced by this handler", mt),
			Location:       location.OfCall(),
		})
	}

	if s.instance.ended {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "ProcessMessageHandler",
			Method:         s.handleMethod,
			Implementation: s.config.Implementation(),
			Message:        s.env.Message,
			Description:    fmt.Sprintf("executed a command of type %s on an ended process", mt),
			Location:       location.OfCall(),
		})
	}

	if err := m.Validate(validation.CommandValidationScope()); err != nil {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "ProcessMessageHandler",
			Method:         s.handleMethod,
			Message:        s.env.Message,
			Implementation: s.config.Implementation(),
			Description:    fmt.Sprintf("executed an invalid %s command: %s", mt, err),
			Location:       location.OfCall(),
		})
	}

	env := s.env.NewCommand(
		s.messageIDs.Next(),
		m,
		s.now,
		envelope.Origin{
			Handler:     s.config,
			HandlerType: config.ProcessHandlerType,
			InstanceID:  s.instanceID,
		},
	)

	s.commands = append(s.commands, env)

	s.observer.Notify(fact.CommandExecutedByProcess{
		Handler:         s.config,
		InstanceID:      s.instanceID,
		Root:            s.root,
		Envelope:        s.env,
		CommandEnvelope: env,
	})
}

func (s *scope) RecordedAt() time.Time {
	s.guardAgainstDirectMutation("RecordedAt", location.OfCall())
	return s.env.CreatedAt
}

func (s *scope) ScheduleDeadline(m dogma.Deadline, t time.Time) {
	s.guardAgainstDirectMutation("ScheduleDeadline", location.OfCall())

	mt := message.TypeOf(m)

	if !s.config.RouteSet().DirectionOf(mt).Has(config.OutboundDirection) {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "ProcessMessageHandler",
			Method:         s.handleMethod,
			Implementation: s.config.Implementation(),
			Message:        s.env.Message,
			Description:    fmt.Sprintf("scheduled a deadline of type %s, which is not produced by this handler", mt),
			Location:       location.OfCall(),
		})
	}

	if s.instance.ended {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "ProcessMessageHandler",
			Method:         s.handleMethod,
			Implementation: s.config.Implementation(),
			Message:        s.env.Message,
			Description:    fmt.Sprintf("scheduled a deadline of type %s on an ended process", mt),
			Location:       location.OfCall(),
		})
	}

	if err := m.Validate(validation.DeadlineValidationScope()); err != nil {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "ProcessMessageHandler",
			Method:         s.handleMethod,
			Message:        s.env.Message,
			Implementation: s.config.Implementation(),
			Description:    fmt.Sprintf("scheduled an invalid %s deadline: %s", mt, err),
			Location:       location.OfCall(),
		})
	}

	env := s.env.NewDeadline(
		s.messageIDs.Next(),
		m,
		s.now,
		t,
		envelope.Origin{
			Handler:     s.config,
			HandlerType: config.ProcessHandlerType,
			InstanceID:  s.instanceID,
		},
	)

	if t.After(s.now) {
		s.pending = append(s.pending, env)
	} else {
		s.ready = append(s.ready, env)
	}

	s.observer.Notify(fact.DeadlineScheduledByProcess{
		Handler:          s.config,
		InstanceID:       s.instanceID,
		Root:             s.root,
		Envelope:         s.env,
		DeadlineEnvelope: env,
	})
}

func (s *scope) ScheduledFor() time.Time {
	s.guardAgainstDirectMutation("ScheduledFor", location.OfCall())
	return s.env.ScheduledFor
}

func (s *scope) Now() time.Time {
	s.guardAgainstDirectMutation("Now", location.OfCall())
	return s.now
}

func (s *scope) Log(f string, v ...any) {
	s.guardAgainstDirectMutation("Log", location.OfCall())

	s.observer.Notify(fact.MessageLoggedByProcess{
		Handler:      s.config,
		InstanceID:   s.instanceID,
		Root:         s.root,
		Ended:        s.instance.ended,
		Envelope:     s.env,
		LogFormat:    f,
		LogArguments: v,
	})
}

func (s *scope) guardAgainstDirectMutation(method string, loc location.Location) {
	thisOp := ""
	if method != "" {
		thisOp = fmt.Sprintf("call to %s() at %s", method, loc)
	}

	if !compare.Equal(s.root, s.shadowRoot) {
		desc := "modified the process root without using Mutate()"

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
			Interface:      "ProcessMessageHandler",
			Method:         s.handleMethod,
			Implementation: s.config.Implementation(),
			Message:        s.env.Message,
			Description:    desc,
			Location:       location.OfCall(),
		})
	}

	s.lastOp = thisOp
}
