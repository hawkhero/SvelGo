import protobuf from 'protobufjs/light'
import descriptor from './ui_descriptor.json'

export const root = protobuf.Root.fromJSON(descriptor as protobuf.INamespace)

const PageStateMsg   = root.lookupType('ui.PageState')
const ClientEventMsg = root.lookupType('ui.ClientEvent')
const StateUpdateMsg = root.lookupType('ui.StateUpdate')

// Map component type → protobuf message type for decoding state_bytes.
// Applications register their component decoders by calling registerComponentDecoder().
const componentTypes: Record<string, protobuf.Type> = {}

export function registerComponentDecoder(typeName: string, msgType: protobuf.Type) {
  componentTypes[typeName] = msgType
}

// Decode the base64 protobuf blob injected into the HTML shell
export function decodePageState(base64blob: string): protobuf.Message {
  const bytes = Uint8Array.from(atob(base64blob), c => c.charCodeAt(0))
  return PageStateMsg.decode(bytes)
}

// Decode a component's state_bytes using its declared type
export function decodeComponentState(type: string, stateBytes: Uint8Array): Record<string, unknown> {
  const MsgType = componentTypes[type]
  if (!MsgType) {
    console.warn(`Unknown component type: ${type}`)
    return {}
  }
  const msg = MsgType.decode(stateBytes)
  return MsgType.toObject(msg, { defaults: true, longs: String, enums: String }) as Record<string, unknown>
}

// Decode a binary WebSocket frame as a StateUpdate
export function decodeStateUpdate(buffer: ArrayBuffer): protobuf.Message {
  return StateUpdateMsg.decode(new Uint8Array(buffer))
}

// Encode a ClientEvent to send over the WebSocket
export function encodeClientEvent(payload: Record<string, unknown>): Uint8Array {
  const msg = ClientEventMsg.create(payload)
  return ClientEventMsg.encode(msg).finish()
}
