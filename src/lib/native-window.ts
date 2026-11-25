import { Events, Window } from '@wailsio/runtime'

export const emit = Events.Emit

export async function listen<T>(
  name: string,
  callback: (event: { payload: T }) => void,
): Promise<() => void> {
  return Events.On(name, ({ data }) => callback({ payload: data as T }))
}

export const currentWindow = {
  show: () => Window.Show(),
  hide: () => Window.Hide(),
  close: () => Window.Close(),
  setFocus: () => Window.Focus(),
  minimize: () => Window.Minimise(),
  unminimize: () => Window.UnMinimise(),
  isMaximized: () => Window.IsMaximised(),
  isFullscreen: () => Window.IsFullscreen(),
  toggleMaximize: () => Window.ToggleMaximise(),
  setFullscreen: (enabled: boolean) =>
    enabled ? Window.Fullscreen() : Window.UnFullscreen(),
  onResized: (update: () => void) =>
    listen(Events.Types.Common.WindowDidResize, update),
  onFocusChanged: async (update: (event: { payload: boolean }) => void) => {
    const focus = Events.On(Events.Types.Common.WindowFocus, () =>
      update({ payload: true }),
    )
    const blur = Events.On(Events.Types.Common.WindowLostFocus, () =>
      update({ payload: false }),
    )
    return () => {
      focus()
      blur()
    }
  },
}
