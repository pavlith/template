import { commands } from '$lib/wails-bindings'
import { Events } from '@wailsio/runtime'
import type { AppPlatform } from '$lib/hooks/use-platform.svelte'
import { getPlatform } from '$lib/hooks/use-platform.svelte'
import { executeCommand } from '$lib/commands/registry.svelte'
import { getCommand, getEffectiveShortcut } from '$lib/commands/registry.svelte'
import { toNativeAccelerator } from '$lib/shortcuts'
import i18n from '$lib/i18n/config'
import {
  TOGGLE_THEME,
  OPEN_PREFERENCES,
  TOGGLE_LEFT_SIDEBAR,
  TOGGLE_RIGHT_SIDEBAR,
  TOGGLE_QUICK_PANE,
  APP_QUIT,
} from '$lib/commands/app-commands'

export interface MenuItemSpec {
  id: string
  labelKey: string
  commandId: string
  accelerator?: string
}

export interface SeparatorSpec {
  separator: true
}

export type PredefinedItemKind =
  | 'Separator'
  | 'Undo'
  | 'Redo'
  | 'Cut'
  | 'Copy'
  | 'Paste'
  | 'SelectAll'
  | 'Minimize'
  | 'Maximize'
  | 'Fullscreen'
  | 'CloseWindow'
  | 'Hide'
  | 'HideOthers'
  | 'ShowAll'
  | 'BringAllToFront'
  | { About: null }

export interface PredefinedSpec {
  predefined: PredefinedItemKind
}

export interface SubmenuSpec {
  labelKey: string
  items: (MenuItemSpec | SeparatorSpec | PredefinedSpec)[]
}

function commandItem(
  commandId: string,
  labelKeyOverride?: string,
): MenuItemSpec {
  const command = getCommand(commandId)
  const shortcut = command ? getEffectiveShortcut(command) : undefined
  return {
    id: commandId,
    labelKey: labelKeyOverride ?? command?.labelKey ?? commandId,
    commandId,
    accelerator: shortcut ? toNativeAccelerator(shortcut) : undefined,
  }
}

function sep(): SeparatorSpec {
  return { separator: true }
}

function predefined(item: PredefinedSpec['predefined']): PredefinedSpec {
  return { predefined: item }
}

/**
 * Submenus that are identical on every platform.
 *
 * They are built fresh per call rather than shared as constants, because
 * `commandItem()` reads the live registry — a command's accelerator can change
 * while the app is running, and the menu is rebuilt on `languageChanged`.
 */
function editSubmenu(): SubmenuSpec {
  return {
    labelKey: 'menu.edit',
    items: [
      predefined('Undo'),
      predefined('Redo'),
      sep(),
      predefined('Cut'),
      predefined('Copy'),
      predefined('Paste'),
      sep(),
      predefined('SelectAll'),
    ],
  }
}

function viewSubmenu(): SubmenuSpec {
  return {
    labelKey: 'menu.view',
    items: [
      commandItem(TOGGLE_LEFT_SIDEBAR),
      commandItem(TOGGLE_RIGHT_SIDEBAR),
      sep(),
      commandItem(TOGGLE_QUICK_PANE),
      commandItem(TOGGLE_THEME),
    ],
  }
}

export function buildMenuSpec(platform: AppPlatform): SubmenuSpec[] {
  if (platform === 'macos') {
    return [
      {
        labelKey: 'app.name',
        items: [
          predefined({ About: null }),
          sep(),
          commandItem(OPEN_PREFERENCES, 'titlebar.settings'),
          sep(),
          predefined('Hide'),
          predefined('HideOthers'),
          predefined('ShowAll'),
          sep(),
          // Custom MenuItem instead of PredefinedMenuItem('Quit')
          // to ensure close-handshake (flush stores) is respected.
          commandItem(APP_QUIT, 'commands.quit'),
        ],
      },
      editSubmenu(),
      viewSubmenu(),
      {
        labelKey: 'menu.window',
        items: [predefined('Minimize'), predefined('Maximize')],
      },
    ]
  }

  // Windows / Linux — menu is built but NOT attached to the window
  // (decorations: false means a native menu strip would conflict
  // with the custom titlebar). Users access these via the command palette.
  return [
    {
      labelKey: 'menu.file',
      items: [
        commandItem(OPEN_PREFERENCES, 'titlebar.settings'),
        sep(),
        commandItem(APP_QUIT, 'commands.quit'),
      ],
    },
    editSubmenu(),
    viewSubmenu(),
  ]
}

async function buildAndApplyMenu(): Promise<void> {
  const platform = getPlatform()
  if (platform !== 'macos') return
  const menu = buildMenuSpec(platform).map((sub) => ({
    label: i18n.t(sub.labelKey),
    items: sub.items.map((item) => {
      if ('commandId' in item) {
        return {
          id: item.commandId,
          label: i18n.t(item.labelKey),
          accelerator: item.accelerator ?? '',
          role: '',
          disabled: false,
        }
      }
      return {
        id: '',
        label: '',
        accelerator: '',
        role:
          'separator' in item
            ? 'Separator'
            : typeof item.predefined === 'string'
              ? item.predefined
              : 'About',
        disabled: false,
      }
    }),
  }))
  await commands.setAppMenu(menu)
}

/**
 * Rebuilds the menu so it reflects registry state that changed after boot —
 * currently called when the user rebinds a command shortcut, mirroring the
 * languageChanged rebuild below.
 */
export async function rebuildMenu(): Promise<void> {
  await buildAndApplyMenu()
}

export async function initMenu(): Promise<() => void> {
  await buildAndApplyMenu()

  const onLanguageChanged = () => {
    void buildAndApplyMenu()
  }
  i18n.on('languageChanged', onLanguageChanged)
  const off = Events.On('app:menu-command', ({ data }) => {
    void executeCommand(data as string)
  })

  return () => {
    off()
    i18n.off('languageChanged', onLanguageChanged)
  }
}
