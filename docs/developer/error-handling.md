# Error Handling

The Svelte ErrorBoundary provides a fallback for failures in the main content.
The frontend logger in src/lib/logger.ts forwards production warnings and
errors to Go; error reports include available stack information. Go panics
are recorded by the main recovery handler and Wails PanicHandler.
Desktop.HasRecentCrash and GetCrashReport support the startup recovery notice.

Generated Wails calls return promises that reject on Go errors. The adapter
in src/lib/wails-bindings.ts wraps selected calls in a discriminated result;
unwrapResult rethrows failures. Handle errors at user-action boundaries and
do not silently replace native operations with mock successes.

## Quit and recovery

The main window hook cancels WindowClosing and emits app:close-requested.
The frontend checks its dirty flag, confirms if needed, flushes preferences
and application state and then hides on macOS or quits elsewhere.
Tray Quit and the explicit quit command use requestQuit in lifecycle.ts.
A failure during the close flush is logged; the existing window-close path
still permits exit. The dirty flag is frontend state, not a Go mirror.

Desktop.SaveEmergencyData, CleanupOldRecoveryFiles, ListCrashReports,
GetCrashReport, ClearCrashReports and CollectDiagnostics provide recovery
and diagnostic support. Review diagnostics before sharing them.
