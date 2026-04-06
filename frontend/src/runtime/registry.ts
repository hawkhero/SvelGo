// Maps component type strings (from Go) to Svelte component constructors.
// Applications register their components by calling registerComponent().
export const ComponentRegistry: Record<string, any> = {}

export function registerComponent(typeName: string, ctor: any) {
  ComponentRegistry[typeName] = ctor
}
