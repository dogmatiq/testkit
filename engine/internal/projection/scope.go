package projection

import (
	"time"

	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
)

// scope is an implementation of [dogma.ProjectionEventScope] and
// [dogma.ProjectionCompactScope].
type scope struct {
	config     *config.Projection
	observer   fact.Observer
	event      *envelope.Envelope // nil if compacting
	checkpoint uint64
	now        time.Time
}

func (s *scope) RecordedAt() time.Time {
	return s.event.CreatedAt
}

func (s *scope) StreamID() string {
	return s.event.EventStreamID
}

func (s *scope) Offset() uint64 {
	return s.event.EventStreamOffset
}

func (s *scope) CheckpointOffset() uint64 {
	return s.checkpoint
}

func (s *scope) Now() time.Time {
	return s.now
}

func (s *scope) Log(f string, v ...any) {
	s.observer.Notify(fact.MessageLoggedByProjection{
		Handler:      s.config,
		Envelope:     s.event, // nil if compacting
		LogFormat:    f,
		LogArguments: v,
	})
}
