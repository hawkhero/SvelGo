<script lang="ts">
  import type { Writable } from 'svelte/store'
  import { getComponentStore } from '../runtime/state'
  import { sendEvent } from '../runtime/ws'

  let { id }: { id: string } = $props()

  // Subscribe to the reactive store so Svelte 5 tracks it properly
  let state: Record<string, unknown> = $state({})
  $effect(() => {
    const store = getComponentStore(id) as Writable<Record<string, unknown>>
    return store.subscribe(s => { state = s })
  })
</script>

<button
  onclick={() => sendEvent(id, 'click')}
  style="
    background-color: #3b82f6;
    color: white;
    padding: 0.75rem 1.5rem;
    border: none;
    border-radius: 0.5rem;
    font-size: 1rem;
    cursor: pointer;
  "
>
  {state.label ?? ''}
</button>
