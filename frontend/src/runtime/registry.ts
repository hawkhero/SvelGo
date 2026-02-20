import Button from '../components/Button.svelte'

// Maps component type strings (from Go) to Svelte component constructors.
// Add new components here as the framework grows.
export const ComponentRegistry: Record<string, any> = {
  Button,
}
