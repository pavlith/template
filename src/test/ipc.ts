import { vi } from 'vitest'

type CommandHandler = (name: string, args?: Record<string, unknown>) => unknown

const mock = vi.hoisted(() => ({
  handler: (() => undefined) as CommandHandler,
}))

vi.mock(
  '../../bindings/github.com/pavlith/template/desktop',
  async (original) => {
    const bindings = await original<Record<string, unknown>>()
    const parameters: Record<string, string[]> = {
      SavePreferences: ['preferences'],
      SaveAppState: ['appState'],
      RegisterGlobalShortcut: ['purpose', 'accelerator'],
      UnregisterGlobalShortcut: ['purpose'],
    }
    return Object.fromEntries(
      Object.keys(bindings).map((method) => [
        method,
        async (...args: unknown[]) =>
          mock.handler(
            method.replace(
              /[A-Z]/g,
              (char, offset: number) =>
                (offset ? '_' : '') + char.toLowerCase(),
            ),
            Object.fromEntries(
              (parameters[method] ?? []).map((name, index) => [
                name,
                args[index],
              ]),
            ),
          ),
      ]),
    )
  },
)

export function mockIPC(handler: CommandHandler): void {
  mock.handler = handler
}

export function clearMocks(): void {
  mock.handler = () => undefined
}
