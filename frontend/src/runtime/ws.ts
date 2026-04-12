import { decodeStateUpdate, decodeComponentState, encodeClientEvent } from './proto'
import { updateComponentState } from './state'

// Typed shape of the decoded StateUpdate protobuf message.
interface DecodedStateUpdate {
  pageId:            string
  updatedComponents: DecodedComponentState[]
}

interface DecodedComponentState {
  id:         string
  type:       string
  stateBytes: Uint8Array
}

const debug = () => (window as unknown as { __SVELGO_DEBUG__: boolean }).__SVELGO_DEBUG__ === true

let socket: WebSocket

export function openWebSocket(pageId: string) {
  const url = `ws://${location.host}/ws?page-id=${encodeURIComponent(pageId)}`
  socket = new WebSocket(url)
  socket.binaryType = 'arraybuffer'

  socket.onmessage = (evt) => {
    const update = decodeStateUpdate(evt.data) as unknown as DecodedStateUpdate
    if (debug()) console.debug('[svelgo ws ←]', update)
    for (const cs of (update.updatedComponents ?? [])) {
      const decoded = decodeComponentState(cs.type, cs.stateBytes)
      updateComponentState(cs.id, decoded)
    }
  }

  socket.onerror = (err) => console.error('SvelGo WebSocket error:', err)
  socket.onclose = () => console.log('SvelGo WebSocket closed')
}

export function sendEvent(componentId: string, eventType: string, payload: Uint8Array = new Uint8Array()) {
  if (!socket || socket.readyState !== WebSocket.OPEN) return
  const pageId = (window as unknown as { __SVELGO_PAGE_ID__: string }).__SVELGO_PAGE_ID__
  const bytes = encodeClientEvent({ pageId, componentId, eventType, payload })
  if (debug()) console.debug('[svelgo ws →]', { pageId, componentId, eventType })
  socket.send(bytes)
}
