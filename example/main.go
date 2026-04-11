package main

import (
	"fmt"
	"log"
	"net/http"

	svelgo "github.com/hawkhero/svelgo"
	"github.com/hawkhero/svelgo/component"
)

func main() {
	svelgo.Setup()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		clickCount := 0

		lbl := &component.Label{
			ID:   "counter-label",
			Text: "Count: 0",
		}

		btn := &component.Button{
			ID:    "counter-btn",
			Label: "Click me (0 clicks)",
		}

		btn.OnClick = func() {
			clickCount++
			btn.Label = fmt.Sprintf("Click me (%d clicks)", clickCount)
			lbl.Text = fmt.Sprintf("Count: %d", clickCount)
			log.Printf("Button clicked — count: %d", clickCount)
		}

		page := svelgo.NewPage()
		page.Add(lbl).Add(btn)
		page.Render(w, r)
	})

	log.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
