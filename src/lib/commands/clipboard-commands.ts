import { Clipboard } from '@wailsio/runtime'
import { toast } from '$lib/stores/toast'
import { t } from '$lib/i18n/t.svelte'
import { logger } from '$lib/logger'
import { registerCommands, type AppCommand } from './registry.svelte'

export const DEMO_COPY_TO_CLIPBOARD = 'demo-copy-to-clipboard'
export const DEMO_PASTE_FROM_CLIPBOARD = 'demo-paste-from-clipboard'

/** Demo for the Wails clipboard API. */
async function copyDemoText(): Promise<void> {
  const text = t('demo.clipboardPayload')
  try {
    await Clipboard.SetText(text)
    toast.success(t('demo.clipboardCopied'), { description: text })
  } catch (e) {
    logger.error('Clipboard write failed', e)
    toast.error(t('demo.clipboardWriteFailed'))
  }
}

async function readClipboard(): Promise<void> {
  try {
    const text = await Clipboard.Text()
    if (text) {
      toast.info(t('demo.clipboardRead'), { description: text })
    } else {
      toast.info(t('demo.clipboardEmpty'))
    }
  } catch (e) {
    logger.error('Clipboard read failed', e)
    toast.error(t('demo.clipboardReadFailed'))
  }
}

const clipboardCommands: AppCommand[] = [
  {
    id: DEMO_COPY_TO_CLIPBOARD,
    labelKey: 'commands.demoCopyToClipboard',
    category: 'commands.category.demo',
    run: copyDemoText,
  },
  {
    id: DEMO_PASTE_FROM_CLIPBOARD,
    labelKey: 'commands.demoPasteFromClipboard',
    category: 'commands.category.demo',
    run: readClipboard,
  },
]

export function registerClipboardCommands(): void {
  registerCommands(clipboardCommands)
}
