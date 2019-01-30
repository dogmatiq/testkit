package render

import (
	"strings"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/iago"
)

// Message returns a human-readable representation of v.
func Message(r Renderer, v dogma.Message) string {
	var w strings.Builder
	iago.Must(r.WriteMessage(&w, v))
	return w.String()
}

// AggregateRoot returns a human-readable representation of v.
func AggregateRoot(r Renderer, v dogma.AggregateRoot) string {
	var w strings.Builder
	iago.Must(r.WriteAggregateRoot(&w, v))
	return w.String()
}

// ProcessRoot returns a human-readable representation of v.
func ProcessRoot(r Renderer, v dogma.ProcessRoot) string {
	var w strings.Builder
	iago.Must(r.WriteProcessRoot(&w, v))
	return w.String()
}

// AggregateMessageHandler returns a human-readable representation of v.
func AggregateMessageHandler(r Renderer, v dogma.AggregateMessageHandler) string {
	var w strings.Builder
	iago.Must(r.WriteAggregateMessageHandler(&w, v))
	return w.String()
}

// ProcessMessageHandler returns a human-readable representation of v.
func ProcessMessageHandler(r Renderer, v dogma.ProcessMessageHandler) string {
	var w strings.Builder
	iago.Must(r.WriteProcessMessageHandler(&w, v))
	return w.String()
}

// IntegrationMessageHandler returns a human-readable representation of v.
func IntegrationMessageHandler(r Renderer, v dogma.IntegrationMessageHandler) string {
	var w strings.Builder
	iago.Must(r.WriteIntegrationMessageHandler(&w, v))
	return w.String()
}

// ProjectionMessageHandler returns a human-readable representation of v.
func ProjectionMessageHandler(r Renderer, v dogma.ProjectionMessageHandler) string {
	var w strings.Builder
	iago.Must(r.WriteProjectionMessageHandler(&w, v))
	return w.String()
}
