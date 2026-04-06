package component

import (
	uipb "github.com/hawkhero/svelgo/gen/ui"
	"google.golang.org/protobuf/proto"
)

// Button is a built-in interactive button component.
// Set OnClick to handle click events in your Go handler.
type Button struct {
	ID       string
	Label    string
	Disabled bool
	OnClick  func()
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

func (b *Button) HandleEvent(eventType string, _ []byte) error {
	if eventType == "click" && b.OnClick != nil {
		b.OnClick()
	}
	return nil
}
