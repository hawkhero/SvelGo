package svelgo

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	uipb "github.com/hawkhero/svelgo/gen/ui"

	"google.golang.org/protobuf/proto"
)

// Page represents a single server-rendered page instance.
type Page struct {
	id         string
	components []Component
}

// NewPage creates a new Page with a unique session ID.
func NewPage() *Page {
	return &Page{id: newID()}
}

// Add appends a component to the page. Returns the page for chaining.
func (p *Page) Add(c Component) *Page {
	p.components = append(p.components, c)
	return p
}

// Render serialises the page state, registers the session, and writes the
// HTML shell response.
func (p *Page) Render(w http.ResponseWriter, r *http.Request) {
	// Build manifest (component list for the JS runtime)
	manifest := make([]ComponentManifestEntry, len(p.components))
	for i, c := range p.components {
		manifest[i] = ComponentManifestEntry{
			ID:   c.ComponentID(),
			Type: c.ComponentType(),
			Slot: c.Slot(),
		}
	}

	// Serialise all component states into a PageState protobuf
	pageState := &uipb.PageState{PageId: p.id}
	for _, c := range p.components {
		stateBytes, err := proto.Marshal(c.ProtoState())
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		pageState.Components = append(pageState.Components, &uipb.ComponentState{
			Id:         c.ComponentID(),
			Type:       c.ComponentType(),
			StateBytes: stateBytes,
		})
	}

	encoded, err := proto.Marshal(pageState)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	stateBlob := base64.StdEncoding.EncodeToString(encoded)

	// Register session for WebSocket event dispatch
	globalSessionStore.Register(p.id, p.components)

	// Resolve asset paths (dev → Vite URLs, prod → hashed bundle)
	scriptPath, cssPath := resolveAssets()

	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := pageRenderData{
		PageID:      p.id,
		StateBlob:   stateBlob,
		Manifest:    template.JS(manifestJSON),
		AssetScript: scriptPath,
		AssetCSS:    cssPath,
		Debug:       debugMode,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := shellTemplate.Execute(w, data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

func newID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
