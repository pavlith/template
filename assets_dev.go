//go:build !production

package main

import (
	"io/fs"
	"os"
)

// Wails proxies the Vite server in development; dist is the local fallback.
var frontendAssets fs.FS = os.DirFS("dist")
