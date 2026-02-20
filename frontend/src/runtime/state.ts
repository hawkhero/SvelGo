import { writable } from 'svelte/store'

type Store = ReturnType<typeof writable<Record<string, unknown>>>

const stores = new Map<string, Store>()

export function initComponentState(id: string, initialState: Record<string, unknown>) {
  stores.set(id, writable(initialState))
}

export function getComponentStore(id: string): Store {
  const store = stores.get(id)
  if (!store) throw new Error(`No state store for component "${id}"`)
  return store
}

export function updateComponentState(id: string, patch: Record<string, unknown>) {
  const store = stores.get(id)
  if (store) {
    store.update(current => ({ ...current, ...patch }))
  }
}
