package process

import (
	"fmt"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine/internal/panicx"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/location"
)

// scope is an implementation of dogma.ProcessEventScope and
// dogma.ProcessTimeoutScope.
type scope struct {
	instanceID   string
	config       configkit.RichProcess
	handleMethod string
	messageIDs   *envelope.MessageIDGenerator
	observer     fact.Observer
	now          time.Time
	root         dogma.ProcessRoot
	ended        bool
	env          *envelope.Envelope // event or timeout
	commands     []*envelope.Envelope
	ready        []*envelope.Envelope // timeouts <= now
	pending      []*envelope.Envelope // timeouts > now
}

func (s *scope) InstanceID() string {
	return s.instanceID
}

func (s *scope) End() {
	if s.ended {
		return
	}

	s.ended = true

	s.observer.Notify(fact.ProcessInstanceEnded{
		Handler:    s.config,
		InstanceID: s.instanceID,
		Root:       s.root,
		Envelope:   s.env,
	})
}

func (s *scope) ExecuteCommand(m dogma.Message) {
	if s.config.MessageTypes().Produced[message.TypeOf(m)] != message.CommandRole {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "ProcessMessageHandler",
			Method:         s.handleMethod,
			Implementation: s.config.Handler(),
			Message:        s.env.Message,
			Description:    fmt.Sprintf("executed a command of type %T, which is not produced by this handler", m),
			Location:       location.OfCall(),
		})
	}

	if err := dogma.ValidateMessage(m); err != nil {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "ProcessMessageHandler",
			Method:         s.handleMethod,
			Message:        s.env.Message,
			Implementation: s.config.Handler(),
			Description:    fmt.Sprintf("executed an invalid %T command: %s", m, err),
			Location:       location.OfCall(),
		})
	}

	if s.ended {
		s.observer.Notify(fact.ProcessInstanceEndingReverted{
			Handler:    s.config,
			InstanceID: s.instanceID,
			Root:       s.root,
			Envelope:   s.env,
		})

		s.ended = false
	}

	env := s.env.NewCommand(
		s.messageIDs.Next(),
		m,
		s.now,
		envelope.Origin{
			Handler:     s.config,
			HandlerType: configkit.ProcessHandlerType,
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
	return s.env.CreatedAt
}

func (s *scope) ScheduleTimeout(m dogma.Message, t time.Time) {
	if s.config.MessageTypes().Produced[message.TypeOf(m)] != message.TimeoutRole {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "ProcessMessageHandler",
			Method:         s.handleMethod,
			Implementation: s.config.Handler(),
			Message:        s.env.Message,
			Description:    fmt.Sprintf("scheduled a timeout of type %T, which is not produced by this handler", m),
			Location:       location.OfCall(),
		})
	}

	if err := dogma.ValidateMessage(m); err != nil {
		panic(panicx.UnexpectedBehavior{
			Handler:        s.config,
			Interface:      "ProcessMessageHandler",
			Method:         s.handleMethod,
			Message:        s.env.Message,
			Implementation: s.config.Handler(),
			Description:    fmt.Sprintf("scheduled an invalid %T timeout: %s", m, err),
			Location:       location.OfCall(),
		})
	}

	if s.ended {
		s.observer.Notify(fact.ProcessInstanceEndingReverted{
			Handler:    s.config,
			InstanceID: s.instanceID,
			Root:       s.root,
			Envelope:   s.env,
		})

		s.ended = false
	}

	env := s.env.NewTimeout(
		s.messageIDs.Next(),
		m,
		s.now,
		t,
		envelope.Origin{
			Handler:     s.config,
			HandlerType: configkit.ProcessHandlerType,
			InstanceID:  s.instanceID,
		},
	)

	if t.After(s.now) {
		s.pending = append(s.pending, env)
	} else {
		s.ready = append(s.ready, env)
	}

	s.observer.Notify(fact.TimeoutScheduledByProcess{
		Handler:         s.config,
		InstanceID:      s.instanceID,
		Root:            s.root,
		Envelope:        s.env,
		TimeoutEnvelope: env,
	})
}

func (s *scope) ScheduledFor() time.Time {
	return s.env.ScheduledFor
}

func (s *scope) Log(f string, v ...interface{}) {
	s.observer.Notify(fact.MessageLoggedByProcess{
		Handler:      s.config,
		InstanceID:   s.instanceID,
		Root:         s.root,
		Envelope:     s.env,
		LogFormat:    f,
		LogArguments: v,
	})
}
