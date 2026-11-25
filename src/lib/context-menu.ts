import { Events } from '@wailsio/runtime'
import { commands } from '$lib/wails-bindings'
import i18n from '$lib/i18n/config'

export interface ContextMenuItem {
  id: string
  labelKey: string
  action: () => void
  disabled?: boolean
}

export interface ContextMenuSeparator {
  separator: true
}

export type ContextMenuEntry = ContextMenuItem | ContextMenuSeparator

export function isSeparator(
  entry: ContextMenuEntry,
): entry is ContextMenuSeparator {
  return 'separator' in entry && entry.separator === true
}

export interface MenuItemOption {
  id: string
  text: string
  enabled: boolean
  action: () => void
}

export interface SeparatorOption {
  item: 'Separator'
}

export type MenuEntryOption = MenuItemOption | SeparatorOption

export function toMenuItemOptions(
  entries: ContextMenuEntry[],
): MenuEntryOption[] {
  return entries.map((entry) => {
    if (isSeparator(entry)) {
      return { item: 'Separator' as const }
    }
    return {
      id: entry.id,
      text: i18n.t(entry.labelKey),
      enabled: !entry.disabled,
      action: entry.action,
    }
  })
}

let request = 0
let stopListening: (() => void) | undefined

async function buildAndShowMenu(
  options: MenuEntryOption[],
  point: Pick<MouseEvent, 'clientX' | 'clientY'>,
): Promise<void> {
  const id = ++request
  stopListening?.()
  stopListening = Events.On('app:context-selection', ({ data }) => {
    const selection = data as { request: number; id: string }
    if (selection.request !== id) return
    stopListening?.()
    stopListening = undefined
    const item = options.find(
      (option): option is MenuItemOption =>
        'id' in option && option.id === selection.id,
    )
    item?.action()
  })
  await commands.showContextMenu(
    id,
    point.clientX,
    point.clientY,
    options.map((item) =>
      'item' in item
        ? {
            id: '',
            label: '',
            role: item.item,
            disabled: false,
            accelerator: '',
          }
        : {
            id: item.id,
            label: item.text,
            role: '',
            disabled: !item.enabled,
            accelerator: '',
          },
    ),
  )
}

/**
 * The OS supplies these; they need no `action` and no translation, because
 * the platform labels them in the user's system language.
 */
type PredefinedItem = 'Undo' | 'Redo' | 'Cut' | 'Copy' | 'Paste' | 'SelectAll'

const EDIT_ITEMS = [
  'Cut',
  'Copy',
  'Paste',
  'Separator',
  'SelectAll',
] as const satisfies readonly (PredefinedItem | 'Separator')[]

const TEXT_INPUT_ITEMS = ['Undo', 'Redo', 'Separator', ...EDIT_ITEMS] as const

async function popupPredefinedMenu(
  kinds: readonly (PredefinedItem | 'Separator')[],
  point: Pick<MouseEvent, 'clientX' | 'clientY'>,
): Promise<void> {
  await commands.showContextMenu(
    0,
    point.clientX,
    point.clientY,
    kinds.map((role) => ({
      id: '',
      label: '',
      role,
      disabled: false,
      accelerator: '',
    })),
  )
}

/**
 * Builds an arbitrary native menu from a declarative list.
 *
 * @public Part of the template's API surface — kept even though nothing in
 * the starter calls it, because a custom right-click menu is the first thing
 * most consumers reach for.
 */
export async function showContextMenu(
  entries: ContextMenuEntry[],
  point: Pick<MouseEvent, 'clientX' | 'clientY'>,
): Promise<void> {
  const options = toMenuItemOptions(entries)
  await buildAndShowMenu(options, point)
}

/**
 * Cut / Copy / Paste / Select All, without the Undo-Redo pair.
 *
 * @public For read-only surfaces where undo history is meaningless. Text
 * inputs should use `showTextInputContextMenu()` instead.
 */
export async function showEditContextMenu(
  point: Pick<MouseEvent, 'clientX' | 'clientY'>,
): Promise<void> {
  await popupPredefinedMenu(EDIT_ITEMS, point)
}

/**
 * The full text-field menu: Undo/Redo on top of the edit items.
 *
 * Applied to plain inputs through the `textInputContextMenu` action in
 * `$lib/actions/context-menu-actions`.
 */
export async function showTextInputContextMenu(
  point: Pick<MouseEvent, 'clientX' | 'clientY'>,
): Promise<void> {
  await popupPredefinedMenu(TEXT_INPUT_ITEMS, point)
}
