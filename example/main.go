package main

import (
	"fmt"
	"log"
	"net/http"

	svelgo "github.com/svelgo/svelgo"
	"github.com/svelgo/svelgo/component"
)

func main() {
	svelgo.Setup()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		clickCount := 0
		btn := &component.Button{
			ID:    "btn-1",
			Label: "Click me (0 clicks)",
		}
		btn.OnClick = func() {
			clickCount++
			btn.Label = fmt.Sprintf("Click me (%d clicks)", clickCount)
			log.Printf("Button %q clicked — count: %d", btn.ID, clickCount)
		}

		page := svelgo.NewPage()
		page.Add(btn)
		page.Render(w, r)
	})

	log.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
