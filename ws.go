package svelgo

import (
	"log"
	"net/http"

	uipb "github.com/svelgo/svelgo/gen/ui"

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

		// Send updated state back to the client
		stateBytes, err := proto.Marshal(comp.ProtoState())
		if err != nil {
			continue
		}

		update := &uipb.StateUpdate{
			PageId: pageID,
			UpdatedComponents: []*uipb.ComponentState{
				{
					Id:         comp.ComponentID(),
					Type:       comp.ComponentType(),
					StateBytes: stateBytes,
				},
			},
		}
		updateBytes, err := proto.Marshal(update)
		if err != nil {
			continue
		}

		sess.mu.Lock()
		if sess.conn != nil {
			sess.conn.WriteMessage(websocket.BinaryMessage, updateBytes)
		}
		sess.mu.Unlock()
	}
}
