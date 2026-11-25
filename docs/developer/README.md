# Developer documentation

The desktop entry point is `main.go`. The `Desktop` service lives in
`desktop.go` and `menus.go`, and reads the existing JSON stores through
`storage.go`. Svelte has separate Vite entry points for the main window and
Quick Pane. Production builds embed both in the Go binary.

| Command | Purpose |
| --- | --- |
| `pnpm desktop:dev` | Launch Wails with Vite hot reload |
| `pnpm bindings` | Regenerate Go–TypeScript bindings in `bindings/` |
| `pnpm check:all` | Run frontend and Go quality checks |
| `pnpm desktop:build` | Build the production desktop binary |

See the [release guide](releases.md) for packaging. The guides below describe the Wails implementation. Platform-dependent
behavior still needs validation on the corresponding operating system.

| Guide | Topics |
| --- | --- |
| [Architecture Guide](architecture-guide.md) | Project layout, boot sequence, dispatch model, Svelte 5 patterns |
| [State Management](state-management.md) | Stores, schemas, debounced persistence, flush-on-close |
| [Persistence & Recovery](persistence-and-recovery.md) | JSON store, atomic writes, corrupt-file recovery, crash data |
| [Commands & Shortcuts](commands-and-shortcuts.md) | Command registry, palette, keyboard shortcuts, global shortcuts, menus |
| [Theme System](theme-system.md) | Light/dark/system modes, flash-free boot, cross-window sync |
| [Internationalisation](i18n.md) | i18next setup, locale management, RTL support, adding languages |
| [Cross-Platform](cross-platform.md) | Platform detection, titlebar variants, context menus, deep links |
| [Quick Panes](quick-panes.md) | Multi-window architecture, Vite dual-entry, cross-window events |
| [Error Handling](error-handling.md) | Error boundary, logging, crash reporting, diagnostics, quit confirmation, two-phase close |
| [Quality](quality.md) | Vitest, Go tests, ast-grep, knip, jscpd, check:all pipeline |
| [External APIs](external-apis.md) | Go net/http vs frontend fetch, auth token storage, offline caching |
| [Bundle Optimisation](bundle-optimization.md) | Production embedding, tree-shaking, dependency review |
