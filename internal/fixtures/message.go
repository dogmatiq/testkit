package fixtures

import (
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
)

// MessageA is a test implementation of dogma.Message.
type MessageA struct {
	Value interface{}
}

// MessageB is a test implementation of dogma.Message.
type MessageB struct {
	Value interface{}
}

// MessageC is a test implementation of dogma.Message.
type MessageC struct {
	Value interface{}
}

// MessageD is a test implementation of dogma.Message.
type MessageD struct {
	Value interface{}
}

// MessageE is a test implementation of dogma.Message.
type MessageE struct {
	Value interface{}
}

// MessageF is a test implementation of dogma.Message.
type MessageF struct {
	Value interface{}
}

// MessageG is a test implementation of dogma.Message.
type MessageG struct {
	Value interface{}
}

var (
	// MessageA1 is a message of type MessageA for use in tests.
	MessageA1 = MessageA{"<A1>"}

	// MessageA2 is a message of type MessageA for use in tests.
	MessageA2 = MessageA{"<A2>"}

	// MessageA3 is a message of type MessageA for use in tests.
	MessageA3 = MessageA{"<A3>"}

	// MessageB1 is a message of type MessageB for use in tests.
	MessageB1 = MessageB{"<B1>"}

	// MessageB2 is a message of type MessageB for use in tests.
	MessageB2 = MessageB{"<B2>"}

	// MessageB3 is a message of type MessageB for use in tests.
	MessageB3 = MessageB{"<B3>"}

	// MessageC1 is a message of type MessageC for use in tests.
	MessageC1 = MessageC{"<C1>"}

	// MessageC2 is a message of type MessageC for use in tests.
	MessageC2 = MessageC{"<C2>"}

	// MessageC3 is a message of type MessageC for use in tests.
	MessageC3 = MessageC{"<C3>"}

	// MessageD1 is a message of type MessageD for use in tests.
	MessageD1 = MessageD{"<D1>"}

	// MessageD2 is a message of type MessageD for use in tests.
	MessageD2 = MessageD{"<D2>"}

	// MessageD3 is a message of type MessageD for use in tests.
	MessageD3 = MessageD{"<D3>"}

	// MessageE1 is a message of type MessageE for use in tests.
	MessageE1 = MessageE{"<E1>"}

	// MessageE2 is a message of type MessageE for use in tests.
	MessageE2 = MessageE{"<E2>"}

	// MessageE3 is a message of type MessageE for use in tests.
	MessageE3 = MessageE{"<E3>"}

	// MessageF1 is a message of type MessageF for use in tests.
	MessageF1 = MessageF{"<F1>"}

	// MessageF2 is a message of type MessageF for use in tests.
	MessageF2 = MessageF{"<F2>"}

	// MessageF3 is a message of type MessageF for use in tests.
	MessageF3 = MessageF{"<F3>"}

	// MessageG1 is a message of type MessageG for use in tests.
	MessageG1 = MessageG{"<G1>"}

	// MessageG2 is a message of type MessageG for use in tests.
	MessageG2 = MessageG{"<G2>"}

	// MessageG3 is a message of type MessageG for use in tests.
	MessageG3 = MessageG{"<G3>"}
)

var (
	// MessageAType is the message.Type for the MessageA message struct.
	MessageAType = message.TypeOf(MessageA{})

	// MessageBType is the message.Type for the MessageB message struct.
	MessageBType = message.TypeOf(MessageB{})

	// MessageCType is the message.Type for the MessageC message struct.
	MessageCType = message.TypeOf(MessageC{})

	// MessageDType is the message.Type for the MessageD message struct.
	MessageDType = message.TypeOf(MessageD{})

	// MessageEType is the message.Type for the MessageE message struct.
	MessageEType = message.TypeOf(MessageE{})

	// MessageFType is the message.Type for the MessageF message struct.
	MessageFType = message.TypeOf(MessageF{})

	// MessageGType is the message.Type for the MessageG message struct.
	MessageGType = message.TypeOf(MessageG{})
)
