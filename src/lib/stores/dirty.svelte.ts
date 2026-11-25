let _dirty = $state(false)

export function getHasUnsavedChanges(): boolean {
  return _dirty
}

export function setHasUnsavedChanges(value: boolean): void {
  _dirty = value
}
