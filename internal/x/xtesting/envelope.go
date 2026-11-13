package xtesting

import (
	"github.com/dogmatiq/testkit/envelope"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// IgnoreMessageID is a comparison option that ignores the MessageID field of
// [envelope.Envelope].
//
// The engine gives each message a random UUID, so a test can't predict the
// MessageID of an envelope that the engine creates. Use this option to compare
// envelopes that the engine creates from an envelope that the test creates.
// The other fields, including CausationID and CorrelationID, are still
// compared.
var IgnoreMessageID = cmpopts.IgnoreFields(envelope.Envelope{}, "MessageID")

// IgnoreEnvelopeIDs is a comparison option that ignores the MessageID,
// CausationID, CorrelationID and EventStreamID fields of [envelope.Envelope].
//
// Use it to compare an envelope that the engine creates for a message that the
// test dispatches. The engine gives that message a random UUID, and sets the
// envelope's other IDs from it, so a test can't predict any of them.
var IgnoreEnvelopeIDs = cmpopts.IgnoreFields(
	envelope.Envelope{},
	"MessageID",
	"CausationID",
	"CorrelationID",
	"EventStreamID",
)
