package svelgo

import (
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
		sess.mu.Lock()
		sess.conn = nil
		sess.mu.Unlock()
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

		if err := handler.HandleEvent(clientEvent.EventType, clientEvent.Payload); err != nil {
			log.Printf("WS event handler error: %v", err)
			continue
		}

		// Collect updated state for ALL components in the session so that
		// cross-component mutations (e.g. a Button's OnClick updating a Label)
		// are pushed to the client in the same frame.
		sess.mu.Lock()
		var updatedComponents []*uipb.ComponentState
		for _, c := range sess.components {
			stateBytes, err := proto.Marshal(c.ProtoState())
			if err != nil {
				log.Printf("WS marshal error for component %s: %v", c.ComponentID(), err)
				continue
			}
			updatedComponents = append(updatedComponents, &uipb.ComponentState{
				Id:         c.ComponentID(),
				Type:       c.ComponentType(),
				StateBytes: stateBytes,
			})
		}

		update := &uipb.StateUpdate{
			PageId:            pageID,
			UpdatedComponents: updatedComponents,
		}
		updateBytes, err := proto.Marshal(update)
		if err != nil {
			sess.mu.Unlock()
			continue
		}

		if sess.conn != nil {
			sess.conn.WriteMessage(websocket.BinaryMessage, updateBytes)
		}
		sess.mu.Unlock()
	}
}
