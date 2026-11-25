import { System } from '@wailsio/runtime'

export type AppPlatform = 'macos' | 'windows' | 'linux'

let cached: AppPlatform | null = null

export function getPlatform(): AppPlatform {
  if (cached === null) {
    if (System.IsMac() || /Mac/.test(navigator.platform)) cached = 'macos'
    else if (System.IsWindows() || /Win/.test(navigator.platform))
      cached = 'windows'
    else cached = 'linux'
  }
  return cached
}

export function __resetPlatformCache(): void {
  cached = null
}
