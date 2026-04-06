package component

import (
	uipb "github.com/svelgo/svelgo/gen/ui"
	"google.golang.org/protobuf/proto"
)

// Label is a built-in static text display component.
// It has no event handler — state is set once at render time.
type Label struct {
	ID   string
	Text string
}

func (l *Label) ComponentID()   string { return l.ID }
func (l *Label) ComponentType() string { return "svelgo.Label" }
func (l *Label) Slot()          string { return "root" }

func (l *Label) ProtoState() proto.Message {
	return &uipb.LabelState{Text: l.Text}
}
