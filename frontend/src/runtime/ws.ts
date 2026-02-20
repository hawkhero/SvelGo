import { decodeStateUpdate, decodeComponentState, encodeClientEvent } from './proto'
import { updateComponentState } from './state'

let socket: WebSocket

export function openWebSocket(pageId: string) {
  const url = `ws://${location.host}/ws?page-id=${encodeURIComponent(pageId)}`
  socket = new WebSocket(url)
  socket.binaryType = 'arraybuffer'

  socket.onmessage = (evt) => {
    const update = decodeStateUpdate(evt.data) as any
    for (const cs of (update.updatedComponents ?? [])) {
      const decoded = decodeComponentState(cs.type, cs.stateBytes as Uint8Array)
      updateComponentState(cs.id, decoded)
    }
  }

  socket.onerror = (err) => console.error('SvelGo WebSocket error:', err)
  socket.onclose = () => console.log('SvelGo WebSocket closed')
}

export function sendEvent(componentId: string, eventType: string, payload: Uint8Array = new Uint8Array()) {
  if (!socket || socket.readyState !== WebSocket.OPEN) return
  const pageId = (window as any).__SVELGO_PAGE_ID__ as string
  const bytes = encodeClientEvent({ pageId, componentId, eventType, payload })
  socket.send(bytes)
}
