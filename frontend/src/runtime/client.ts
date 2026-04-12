import { mount } from 'svelte'
import { ComponentRegistry } from './registry'
import { initComponentState } from './state'
import { decodePageState, decodeComponentState } from './proto'
import { openWebSocket } from './ws'
import './builtins'

interface ManifestEntry {
  id:   string
  type: string
  slot: string
}

// Typed shape of the decoded PageState protobuf message.
interface DecodedPageState {
  pageId:     string
  components: DecodedComponentState[]
}

interface DecodedComponentState {
  id:         string
  type:       string
  stateBytes: Uint8Array
}

// Typed shape of the globals injected by the Go HTML shell.
interface SvelGoWindow {
  __SVELGO_PAGE_ID__:  string
  __SVELGO_STATE__:    string
  __SVELGO_MANIFEST__: ManifestEntry[]
  __SVELGO_DEBUG__:    boolean
}

export function bootstrap() {
  const w = window as unknown as SvelGoWindow
  const pageId    = w.__SVELGO_PAGE_ID__
  const stateBlob = w.__SVELGO_STATE__
  const manifest  = w.__SVELGO_MANIFEST__

  // Decode the protobuf state blob and initialise per-component Svelte stores
  if (stateBlob) {
    const pageState = decodePageState(stateBlob) as unknown as DecodedPageState
    if (w.__SVELGO_DEBUG__) {
      console.debug('[svelgo init] page state', pageState)
    }
    for (const cs of (pageState.components ?? [])) {
      const decoded = decodeComponentState(cs.type, cs.stateBytes)
      initComponentState(cs.id, decoded)
    }
  }

  // Mount each Svelte component into the root div
  const root = document.getElementById('svelgo-root')!
  for (const entry of manifest) {
    const Ctor = ComponentRegistry[entry.type]
    if (!Ctor) {
      console.warn(`SvelGo: unknown component type "${entry.type}"`)
      continue
    }
    const target = document.createElement('div')
    target.dataset.svelgoSlot = entry.id
    root.appendChild(target)
    mount(Ctor, { target, props: { id: entry.id } })
  }

  // Open the WebSocket for live state updates
  openWebSocket(pageId)
}
