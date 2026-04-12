package svelgo

import (
	"fmt"
	"log"
	"net/http"

	uipb "github.com/hawkhero/svelgo/gen/ui"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WSHandler is the WebSocket endpoint. Register it at "/ws" in your app.
func WSHandler(w http.ResponseWriter, r *http.Request) {
	pageID := r.URL.Query().Get("page-id")
	sess, ok := globalSessionStore.Get(pageID)
	if !ok {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WS upgrade error:", err)
		return
	}
	defer conn.Close()

	sess.mu.Lock()
	sess.conn = conn
	sess.mu.Unlock()

	defer func() {
		// Clean up the connection reference and remove the session so that the
		// component tree is not retained in memory after the browser disconnects.
		sess.mu.Lock()
		sess.conn = nil
		sess.mu.Unlock()
		globalSessionStore.Delete(pageID)
	}()

	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if msgType != websocket.BinaryMessage {
			continue
		}

		clientEvent := &uipb.ClientEvent{}
		if err := proto.Unmarshal(data, clientEvent); err != nil {
			log.Println("WS unmarshal error:", err)
			continue
		}

		sess.mu.Lock()
		comp, ok := sess.components[clientEvent.ComponentId]
		sess.mu.Unlock()

		if !ok {
			continue
		}

		handler, ok := comp.(EventHandler)
		if !ok {
			continue
		}

		if err := handler.HandleEvent(r.Context(), clientEvent.EventType, clientEvent.Payload); err != nil {
			// Include full context so developers can locate the failure immediately.
			log.Printf("WS HandleEvent error: page=%s component=%s type=%s event=%s: %v",
				pageID, clientEvent.ComponentId, comp.ComponentType(), clientEvent.EventType, err)
			continue
		}

		// Collect updated state for ALL components in the session so that
		// cross-component mutations (e.g. a Button's OnClick updating a Label)
		// are pushed to the client in the same frame.
		//
		// If any component fails to marshal we abort the entire update rather
		// than sending a partial StateUpdate that would leave the client in a
		// diverged state.
		sess.mu.Lock()
		updatedComponents, marshalErr := marshalAllComponents(pageID, sess.components)
		if marshalErr != nil {
			sess.mu.Unlock()
			log.Printf("WS state marshal failed (no update sent): page=%s: %v", pageID, marshalErr)
			continue
		}

		update := &uipb.StateUpdate{
			PageId:            pageID,
			UpdatedComponents: updatedComponents,
		}
		updateBytes, err := proto.Marshal(update)
		if err != nil {
			sess.mu.Unlock()
			log.Printf("WS StateUpdate marshal error: page=%s: %v", pageID, err)
			continue
		}

		if sess.conn != nil {
			if writeErr := sess.conn.WriteMessage(websocket.BinaryMessage, updateBytes); writeErr != nil {
				log.Printf("WS write error: page=%s: %v", pageID, writeErr)
				sess.conn = nil
			}
		}
		sess.mu.Unlock()
	}
}

// marshalAllComponents serializes every component's ProtoState into a slice of
// ComponentState messages. It returns an error (and a nil slice) if any single
// component fails to marshal, so callers can enforce all-or-nothing updates.
func marshalAllComponents(pageID string, components map[string]Component) ([]*uipb.ComponentState, error) {
	result := make([]*uipb.ComponentState, 0, len(components))
	for _, c := range components {
		stateBytes, err := proto.Marshal(c.ProtoState())
		if err != nil {
			return nil, fmt.Errorf("component %s (type %s): %w", c.ComponentID(), c.ComponentType(), err)
		}
		result = append(result, &uipb.ComponentState{
			Id:         c.ComponentID(),
			Type:       c.ComponentType(),
			StateBytes: stateBytes,
		})
	}
	return result, nil
}
