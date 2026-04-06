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

export function bootstrap() {
  const pageId   = (window as any).__SVELGO_PAGE_ID__ as string
  const stateBlob = (window as any).__SVELGO_STATE__ as string
  const manifest  = (window as any).__SVELGO_MANIFEST__ as ManifestEntry[]

  // Decode the protobuf state blob and initialise per-component Svelte stores
  if (stateBlob) {
    const pageState = decodePageState(stateBlob) as any
    if ((window as any).__SVELGO_DEBUG__) {
      console.debug('[svelgo init] page state', pageState)
    }
    for (const cs of (pageState.components ?? [])) {
      const decoded = decodeComponentState(cs.type, cs.stateBytes as Uint8Array)
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
