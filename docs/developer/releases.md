# Releases

Run `pnpm release vX.Y.Z` to check the project, synchronize `package.json` and
`wails.yaml`, and prompt before committing, tagging, and pushing.

The release workflow builds Debian packages on Linux, DMGs on macOS, and NSIS
installers on Windows. It attaches the packages to a draft GitHub release.
Use `pnpm wails build --formats deb`, `--formats dmg`, or `--formats nsis`
on the corresponding host to package locally.

For in-app updates, configure `main.updateFeedURL` and `main.updatePublicKey`
through `build.ldflags` in `wails.yaml`. The public key uses base64 encoding.
Publish a compatible Wails endpoint manifest and signed update artifacts.
The About pane enables update checks once the updater has initialized.
