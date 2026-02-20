package svelgo

import "google.golang.org/protobuf/proto"

// Component is the interface every UI component must implement.
type Component interface {
	ComponentID()   string
	ComponentType() string
	Slot()          string
	ProtoState()    proto.Message
}

// EventHandler is an optional interface for components that handle user events.
type EventHandler interface {
	Component
	HandleEvent(eventType string, payload []byte) error
}

// ComponentManifestEntry is serialised as JSON into the HTML shell so the
// browser runtime knows which Svelte component to mount and where.
type ComponentManifestEntry struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Slot string `json:"slot"`
}
