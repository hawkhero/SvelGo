<script lang="ts">
  import type { Writable } from 'svelte/store'
  import { getComponentStore } from '../runtime/state'

  let { id }: { id: string } = $props()

  let state: Record<string, unknown> = $state({})
  $effect(() => {
    const store = getComponentStore(id) as Writable<Record<string, unknown>>
    return store.subscribe(s => { state = s })
  })
</script>

<span style="font-size: 1rem; color: #1f2937;">
  {state.text ?? ''}
</span>
