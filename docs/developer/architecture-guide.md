# Architecture Guide

The template is a single Go module and a Svelte 5 / TypeScript frontend.
Wails owns the native process and webviews; Vite builds two frontend entries.

| Path | Responsibility |
| --- | --- |
| `main.go` | Bootstrap, main and Quick Pane windows, single instance, lifecycle |
| `desktop.go` | Native service methods, shortcuts, updater, diagnostics |
| `storage.go` | JSON persistence, defaults, validation, recovery |
| `menus.go` | Native app and context menus |
| `wails.yaml` | Wake build, frontend, packaging and protocol configuration |
| `bindings/` | Generated Go–TypeScript bindings; do not edit manually |
| `src/lib/wails-bindings.ts` | Typed frontend IPC adapter |
| `src/lib/native-window.ts` | Runtime window and event adapter |
| `src/lib/stores/` | Svelte rune-based state |
| `src/lib/commands/` | Command registry and shortcuts |
| `src/lib/components/` | Layout, dialogs, UI primitives and demo |

## Startup and shutdown

Go creates a hidden main window and a hidden Quick Pane. The frontend
installs lifecycle listeners, restores state, initializes theme, language
and commands, calls MainReady, and shows the main window. Deep links
received before readiness are queued. Saved global shortcuts are registered
on ApplicationStarted.

The main-window close hook cancels native closure and requests a frontend
confirmation and persistence flush. On macOS, window close hides the window;
explicit Quit ends the process. See [Error Handling](error-handling.md).

## IPC and state

Add exported service methods to Desktop, regenerate with `pnpm bindings`,
and use the typed adapter in frontend code. Use Svelte 5 runes and snapshot
reactive data before persistence. Keep domain logic out of UI primitives.
The two webviews have independent JS state; communicate with runtime events
and persisted data as described in [Quick Panes](quick-panes.md).
