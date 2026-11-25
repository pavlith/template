# Persistence and Recovery

`internal/storage` owns JSON persistence and file helpers. Root `storage.go`
keeps the binding models, defaults and validation. `preferences.go` and
`diagnostics.go` expose the corresponding typed methods to the frontend.

Preferences and application state are stored in preferences.json and
state.json. Main-window geometry is stored in window.json. On Linux the
application directory is under XDG_DATA_HOME (default ~/.local/share);
elsewhere it is under the OS user configuration directory. The directory
name is the application ID, com.pavlith.template.

Reads merge JSON into defaults; missing files return defaults, invalid JSON
is renamed to a timestamped .corrupt-* backup. Writes serialize under a mutex
and use a synced temporary file followed by rename. Preference saves validate
theme mode, colors, contrast, font size and other fields. Frontend schema
checks and debounced persistence live in src/lib/stores/.

Recovery snapshots go into recovery/ and crash logs into crash-reports/.
Recovery filenames and input file sizes are constrained. Startup cleans
old recovery files and checks for a recent crash. Diagnostics are collected
through Desktop.CollectDiagnostics.

The new application ID does not automatically migrate the old directory;
see [Using This Template](../../USING_THIS_TEMPLATE.md). Preserve JSON keys
when changing persisted models and add tests for default/legacy data.
