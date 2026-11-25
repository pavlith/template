# Quick Panes

The Quick Pane demonstrates multi-window Wails integration. It is a hidden,
frameless, always-on-top webview created once in main.go, with the URL
/quick-pane.html. Vite builds that HTML entry alongside index.html.

Each window has separate JavaScript state. QuickPaneApp.svelte emits a
submission event; src/lib/quick-pane/bridge.ts listens in the main window,
updates its last-entry state, and displays a toast. Theme hints and theme
events keep the secondary window visually consistent. Listener cleanup is
required when a component unmounts.

Desktop.ShowQuickPane positions the window relative to the main screen where
the OS allows it; ToggleQuickPane and DismissQuickPane control visibility.
The pane hides on close or lost focus. Enter submits and Escape dismisses.
The saved Quick Pane shortcut is registered at application startup.

Use the Go visibility methods rather than creating another native window on
every shortcut press. Keep native methods narrowly scoped and do not assume
a secondary webview is a separate permission boundary.

Test both HTML entries in development and production; confirm the shortcut
opens the pane, submission reaches the main window exactly once, Escape and
focus loss dismiss it, and theme changes are reflected on the next show.
