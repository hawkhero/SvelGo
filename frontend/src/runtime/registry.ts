import type { ComponentType } from 'svelte'

// Maps component type strings (from Go) to Svelte component constructors.
// Applications register their components by calling registerComponent().
export const ComponentRegistry: Record<string, ComponentType> = {}

export function registerComponent(typeName: string, ctor: ComponentType) {
  if (ComponentRegistry[typeName] && ComponentRegistry[typeName] !== ctor) {
    console.warn(
      `SvelGo: registerComponent("${typeName}") is overwriting an existing registration. ` +
      `This is likely a mistake — each type name should be registered exactly once.`
    )
  }
  ComponentRegistry[typeName] = ctor
}
