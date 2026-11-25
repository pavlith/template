import { describe, it, expect, vi } from 'vitest'
import { mockIPC } from '$lib/../test/ipc'

vi.mock('$lib/logger', () => ({
  logger: { warn: vi.fn(), error: vi.fn() },
}))

import { commitAutostart, readAutostartState } from './autostart-toggle'

describe('commitAutostart', () => {
  it('calls the matching plugin command for each direction', async () => {
    const calls: string[] = []
    mockIPC((cmd) => {
      calls.push(cmd)
      return undefined
    })

    expect(await commitAutostart(true)).toEqual({ ok: true })
    expect(await commitAutostart(false)).toEqual({ ok: true })
    expect(calls).toEqual(['enable_autostart', 'disable_autostart'])
  })

  it('reports failure instead of throwing', async () => {
    mockIPC(() => {
      throw new Error('registry says no')
    })

    const result = await commitAutostart(true)
    expect(result.ok).toBe(false)
    if (!result.ok) expect(result.message).toContain('registry says no')
  })
})

describe('readAutostartState', () => {
  it('returns the OS registration state', async () => {
    mockIPC((cmd) => (cmd === 'is_autostart_enabled' ? true : undefined))
    expect(await readAutostartState()).toBe(true)
  })

  it('degrades to false when the plugin call fails', async () => {
    mockIPC(() => {
      throw new Error('no backend')
    })
    expect(await readAutostartState()).toBe(false)
  })
})
