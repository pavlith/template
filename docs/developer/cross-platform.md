# Cross-platform desktop integration

Platform detection is in `src/lib/hooks/use-platform.svelte.ts`. The
platform-specific titlebars live in `src/lib/components/layout/`.
Linux uses native window decorations; Windows and macOS use platform-specific
custom titlebar variants. Window options are set in `main.go`.

## Native operations

Use `src/lib/native-window.ts` for window operations and event subscriptions,
and `src/lib/wails-bindings.ts` for Go service methods. Menus are built in
`menus.go`; the native application menu is installed on macOS. Global
shortcuts, notifications and autostart use Wails facilities. Platform support
must be checked at runtime; in particular Wayland may limit shortcuts and
window positioning. Do not expose unrestricted filesystem/shell operations
through bindings.

## Deep links and tray

`wails-app://` is declared in `wails.yaml` for packaged installations.
`main.go` accepts startup arguments, Wails URL-launch events and
second-instance arguments; queued URLs are returned by MainReady.
Only the configured scheme is accepted. On macOS, test second-instance URL
launch separately; OS launch-event delivery is not equivalent to argv.

The tray Show action restores the main window and Quit uses the frontend
confirmation and persistence path. The Quick Pane has a separate webview.

## Validation

Test native titlebars, focus, minimize/maximize, close/quit, tray, shortcuts,
dialogs, notifications, autostart and deep links on each target OS. Linux is
the first validation target; Windows and macOS remain build targets.
