package aggregate

import (
	"fmt"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine/controller"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
)

// scope is an implementation of dogma.AggregateCommandScope.
type scope struct {
	instanceID string
	config     configkit.RichAggregate
	messageIDs *envelope.MessageIDGenerator
	observer   fact.Observer
	root       dogma.AggregateRoot
	now        time.Time
	exists     bool
	produced   message.TypeCollection
	command    *envelope.Envelope
	events     []*envelope.Envelope
}

func (s *scope) InstanceID() string {
	return s.instanceID
}

func (s *scope) Destroy() {
	if !s.exists {
		return
	}

	s.root = s.config.Handler().New()
	s.exists = false

	s.observer.Notify(fact.AggregateInstanceDestroyed{
		HandlerName: s.config.Identity().Name,
		Handler:     s.config.Handler(),
		InstanceID:  s.instanceID,
		Root:        s.root,
		Envelope:    s.command,
	})
}

func (s *scope) RecordEvent(m dogma.Message) {
	if !s.produced.HasM(m) {
		panic(fmt.Sprintf(
			"the '%s' handler is not configured to record events of type %T",
			s.config.Identity().Name,
			m,
		))
	}

	if err := dogma.ValidateMessage(m); err != nil {
		panic(fmt.Sprintf(
			"can not record event of type %T, it is invalid: %s",
			m,
			err,
		))
	}

	if !s.exists {
		s.observer.Notify(fact.AggregateInstanceCreated{
			HandlerName: s.config.Identity().Name,
			Handler:     s.config.Handler(),
			InstanceID:  s.instanceID,
			Root:        s.root,
			Envelope:    s.command,
		})

		s.exists = true
	}

	controller.ConvertUnexpectedMessagePanic(
		s.config,
		"AggregateRoot",
		"ApplyEvent",
		m,
		func() {
			s.root.ApplyEvent(m)
		},
	)

	env := s.command.NewEvent(
		s.messageIDs.Next(),
		m,
		s.now,
		envelope.Origin{
			HandlerName: s.config.Identity().Name,
			HandlerType: configkit.AggregateHandlerType,
			InstanceID:  s.instanceID,
		},
	)

	s.events = append(s.events, env)

	s.observer.Notify(fact.EventRecordedByAggregate{
		HandlerName:   s.config.Identity().Name,
		Handler:       s.config.Handler(),
		InstanceID:    s.instanceID,
		Root:          s.root,
		Envelope:      s.command,
		EventEnvelope: env,
	})
}

func (s *scope) Log(f string, v ...interface{}) {
	s.observer.Notify(fact.MessageLoggedByAggregate{
		HandlerName:  s.config.Identity().Name,
		Handler:      s.config.Handler(),
		InstanceID:   s.instanceID,
		Root:         s.root,
		Envelope:     s.command,
		LogFormat:    f,
		LogArguments: v,
	})
}
