import { commands } from '$lib/wails-bindings'
import { toast } from '$lib/stores/toast'
import { t } from '$lib/i18n/t.svelte'
import { logger } from '$lib/logger'
import { registerCommands, type AppCommand } from './registry.svelte'

export const DEMO_RUN_SHELL_COMMAND = 'demo-run-shell-command'

async function runShellCommand(): Promise<void> {
  try {
    const output = await commands.runShellDemo()
    toast.success(t('demo.shellRan'), { description: output })
  } catch (e) {
    logger.error('Shell command failed', e)
    toast.error(t('demo.shellFailed'))
  }
}

const shellCommands: AppCommand[] = [
  {
    id: DEMO_RUN_SHELL_COMMAND,
    labelKey: 'commands.demoRunShellCommand',
    category: 'commands.category.demo',
    run: runShellCommand,
  },
]

export function registerShellCommands(): void {
  registerCommands(shellCommands)
}
