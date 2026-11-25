//go:build !production

package main

import "io/fs"

// Wails serves the Vite dev server in development mode.
var frontendAssets fs.FS
