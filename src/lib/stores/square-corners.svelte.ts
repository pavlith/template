import { currentWindow } from '$lib/native-window'
import { getPlatform } from '$lib/hooks/use-platform.svelte'
import { setSquareCorners } from '$lib/stores/ui.svelte'

export function initSquareCorners(): () => void {
  if (getPlatform() === 'macos') return () => {}

  const win = currentWindow
  let cancelled = false

  const update = async () => {
    const isFullscreen = await win.isFullscreen()
    if (cancelled) return
    setSquareCorners(isFullscreen)
  }

  void update()
  const unlistenPromise = win.onResized(() => void update())

  return () => {
    cancelled = true
    void unlistenPromise.then((fn) => fn())
  }
}
