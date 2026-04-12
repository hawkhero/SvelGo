package svelgo

import (
	"context"

	"google.golang.org/protobuf/proto"
)

// Component is the interface every UI component must implement.
type Component interface {
	ComponentID()   string
	ComponentType() string
	Slot()          string
	ProtoState()    proto.Message
}

// EventHandler is an optional interface for components that handle user events.
// The context carries the WebSocket request's cancellation signal, deadline,
// and any request-scoped values (tracing, auth). Pass it to every downstream
// call that accepts a context — database queries, RPCs, etc.
type EventHandler interface {
	Component
	HandleEvent(ctx context.Context, eventType string, payload []byte) error
}

// ComponentManifestEntry is serialised as JSON into the HTML shell so the
// browser runtime knows which Svelte component to mount and where.
type ComponentManifestEntry struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Slot string `json:"slot"`
}
