package main

import (
	"fmt"
	"log"
	"net/http"

	"svelgo"
	uipb "svelgo/gen/ui"

	"google.golang.org/protobuf/proto"
)

// Button is a server-side component. It holds all state and handles events.
type Button struct {
	id         string
	label      string
	clickCount int
}

func (b *Button) ComponentID()   string { return b.id }
func (b *Button) ComponentType() string { return "Button" }
func (b *Button) Slot()          string { return "root" }

func (b *Button) ProtoState() proto.Message {
	return &uipb.ButtonState{
		Label:      b.label,
		ClickCount: int32(b.clickCount),
	}
}

func (b *Button) HandleEvent(eventType string, _ []byte) error {
	if eventType == "click" {
		b.clickCount++
		b.label = fmt.Sprintf("Click me (%d clicks)", b.clickCount)
		log.Printf("Button %q clicked — count: %d", b.id, b.clickCount)
	}
	return nil
}

func main() {
	svelgo.Setup()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		page := svelgo.NewPage()
		page.Add(&Button{
			id:    "btn-1",
			label: "Click me (0 clicks)",
		})
		page.Render(w, r)
	})

	log.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
