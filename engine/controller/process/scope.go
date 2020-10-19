package process

import (
	"fmt"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
)

// scope is an implementation of dogma.ProcessEventScope and
// dogma.ProcessTimeoutScope.
type scope struct {
	instanceID string
	identity   configkit.Identity
	handler    dogma.ProcessMessageHandler
	messageIDs *envelope.MessageIDGenerator
	observer   fact.Observer
	now        time.Time
	root       dogma.ProcessRoot
	exists     bool
	produced   message.TypeCollection
	env        *envelope.Envelope // event or timeout
	commands   []*envelope.Envelope
	ready      []*envelope.Envelope // timeouts <= now
	pending    []*envelope.Envelope // timeouts > now
}

func (s *scope) InstanceID() string {
	return s.instanceID
}

func (s *scope) HasBegun() bool {
	return s.exists
}

func (s *scope) Begin() bool {
	if s.exists {
		return false
	}

	s.exists = true

	s.observer.Notify(fact.ProcessInstanceBegun{
		HandlerName: s.identity.Name,
		Handler:     s.handler,
		InstanceID:  s.instanceID,
		Root:        s.root,
		Envelope:    s.env,
	})

	return true
}

func (s *scope) End() {
	if !s.exists {
		panic("can not end non-existent instance")
	}

	s.exists = false
	s.ready = nil
	s.pending = nil

	s.observer.Notify(fact.ProcessInstanceEnded{
		HandlerName: s.identity.Name,
		Handler:     s.handler,
		InstanceID:  s.instanceID,
		Root:        s.root,
		Envelope:    s.env,
	})
}

func (s *scope) Root() dogma.ProcessRoot {
	if !s.exists {
		panic("can not access process root of non-existent instance")
	}

	return s.root
}

func (s *scope) ExecuteCommand(m dogma.Message) {
	if !s.exists {
		panic("can not execute command against non-existent instance")
	}

	if !s.produced.HasM(m) {
		panic(fmt.Sprintf(
			"the '%s' handler is not configured to execute commands of type %T",
			s.identity.Name,
			m,
		))
	}

	env := s.env.NewCommand(
		s.messageIDs.Next(),
		m,
		s.now,
		envelope.Origin{
			HandlerName: s.identity.Name,
			HandlerType: configkit.ProcessHandlerType,
			InstanceID:  s.instanceID,
		},
	)

	s.commands = append(s.commands, env)

	s.observer.Notify(fact.CommandExecutedByProcess{
		HandlerName:     s.identity.Name,
		Handler:         s.handler,
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
	if !s.exists {
		panic("can not schedule timeout against non-existent instance")
	}

	env := s.env.NewTimeout(
		s.messageIDs.Next(),
		m,
		s.now,
		t,
		envelope.Origin{
			HandlerName: s.identity.Name,
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
		HandlerName:     s.identity.Name,
		Handler:         s.handler,
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
		HandlerName:  s.identity.Name,
		Handler:      s.handler,
		InstanceID:   s.instanceID,
		Root:         s.root,
		Envelope:     s.env,
		LogFormat:    f,
		LogArguments: v,
	})
}
