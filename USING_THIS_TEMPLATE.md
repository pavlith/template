# Using This Template

Install Go 1.26 or later, Node.js 24 or later, and pnpm 11.20.0. On Ubuntu or
Debian, install `libgtk-4-dev`, `libwebkitgtk-6.0-dev`, and `libx11-dev`.

```bash
pnpm install --frozen-lockfile
pnpm desktop:dev
```

The scripts use the Wails CLI pinned in `go.mod`. The development server
listens on `127.0.0.1:1420`.

| Change | Files |
| --- | --- |
| Product name and packaging | `wails.yaml`, `main.go` |
| App ID and data directory | `wails.yaml`, `storage.go` |
| Deep-link protocol | `wails.yaml`, `main.go` |
| Display strings | `locales/` |
| App and tray icon | `build/appicon.png` |

The default identifier is `com.pavlith.template` and the protocol is
`wails-app://`. Choose unique values for your own application. This migration
changes the data directory and protocol from the previous template: existing
installations are not migrated automatically. If needed, copy your old
preferences and state JSON files into the new data directory while the app is
closed. The JSON field names remain compatible.

Add Go methods to the `Desktop` service and regenerate TypeScript with
`pnpm bindings`. The Svelte components use `src/lib/wails-bindings.ts`.
Run `pnpm check:all` and `pnpm desktop:build` before packaging. See the
[release guide](docs/developer/releases.md) for installers and updates.
