# Bundle Optimization

Vite produces dist/ with both index.html and quick-pane.html. The production
Go build embeds dist/ via assets_production.go; development uses the Wails
Vite proxy. Use pnpm desktop:build rather than distributing the dev binary.

Prefer direct imports, keep frontend dependencies lean and inspect Vite's
build output for large chunks. Run knip and jscpd before removing code; keep
vendored UI primitives and generated bindings managed by their respective
generators. Remove unused Go dependencies with go mod tidy after removing
their imports.

Use Wake build configuration for supported compiler flags and packaging.
Measure binary and installer sizes on the target platform rather than assuming
a flag or compression setting applies to every OS. Verify both windows still
load after any bundle change.
