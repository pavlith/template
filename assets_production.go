//go:build production

package main

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var embeddedFrontend embed.FS

var frontendAssets fs.FS = embeddedFrontend
