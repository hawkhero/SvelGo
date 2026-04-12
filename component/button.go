package component

import (
	"context"

	uipb "github.com/hawkhero/svelgo/gen/ui"
	"google.golang.org/protobuf/proto"
)

// Button is a built-in interactive button component.
// Set OnClick to handle click events in your Go handler.
//
// OnClick receives the context from the WebSocket request. Pass it to any
// downstream calls that accept a context (database queries, RPCs, etc.).
// Return a non-nil error to signal that the handler failed; the framework
// will log the error with full context (page ID, component ID, event type)
// and skip the state update for this event.
type Button struct {
	ID       string
	Label    string
	Disabled bool
	OnClick  func(ctx context.Context) error
}

func (b *Button) ComponentID()   string { return b.ID }
func (b *Button) ComponentType() string { return "svelgo.Button" }
func (b *Button) Slot()          string { return "root" }

func (b *Button) ProtoState() proto.Message {
	return &uipb.ButtonState{
		Label:    b.Label,
		Disabled: b.Disabled,
	}
}

func (b *Button) HandleEvent(ctx context.Context, eventType string, _ []byte) error {
	if eventType == "click" && b.OnClick != nil {
		return b.OnClick(ctx)
	}
	return nil
}
