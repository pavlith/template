import { commands } from '$lib/wails-bindings'
import { logger } from '$lib/logger'
import { toast } from '$lib/stores/toast'
import { t } from '$lib/i18n/t.svelte'
import { registerCommands, type AppCommand } from './registry.svelte'

export const DEMO_SEND_NOTIFICATION = 'demo-send-notification'

/**
 * Demo for Wails notifications.
 *
 * Notifications need an OS-level permission that has to be requested at least
 * once, so the real work here is the permission dance rather than the send.
 */
async function sendDemoNotification(): Promise<void> {
  try {
    if (
      !(await commands.notify(
        t('demo.notificationTitle'),
        t('demo.notificationBody'),
      ))
    ) {
      toast.error(t('demo.notificationDenied'))
    }
  } catch (error) {
    logger.error('Notification failed', error)
    toast.error(t('demo.notificationDenied'))
  }
}

const notificationCommands: AppCommand[] = [
  {
    id: DEMO_SEND_NOTIFICATION,
    labelKey: 'commands.demoSendNotification',
    category: 'commands.category.demo',
    run: sendDemoNotification,
  },
]

export function registerNotificationCommands(): void {
  registerCommands(notificationCommands)
}
