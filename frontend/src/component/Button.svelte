<script lang="ts">
  import type { Writable } from 'svelte/store'
  import { getComponentStore } from '../runtime/state'
  import { sendEvent } from '../runtime/ws'

  let { id }: { id: string } = $props()

  let state: Record<string, unknown> = $state({})
  $effect(() => {
    const store = getComponentStore(id) as Writable<Record<string, unknown>>
    return store.subscribe(s => { state = s })
  })
</script>

<button
  onclick={() => sendEvent(id, 'click')}
  disabled={!!state.disabled}
  style="
    background-color: {state.disabled ? '#9ca3af' : '#3b82f6'};
    color: white;
    padding: 0.75rem 1.5rem;
    border: none;
    border-radius: 0.5rem;
    font-size: 1rem;
    cursor: {state.disabled ? 'not-allowed' : 'pointer'};
    opacity: {state.disabled ? '0.6' : '1'};
    transition: background-color 0.15s ease;
  "
>
  {state.label ?? ''}
</button>
