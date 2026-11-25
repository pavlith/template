package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/pavlith/template/internal/storage"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
)

//go:embed build/appicon.png
var appIcon []byte

// package.json also supplies the frontend and distribution version.
//
//go:embed package.json
var packageJSON []byte

const appName = "Wails App"
const deepLinkScheme = "wails-app"

func main() {
	dir, err := storage.DataDirectory(appID)
	if err != nil {
		log.Fatal(err)
	}
	var pkg struct{ Version string }
	if err := json.Unmarshal(packageJSON, &pkg); err != nil {
		log.Fatal(err)
	}
	d := &Desktop{
		store: storage.New(dir), notifier: notifications.New(), version: pkg.Version,
		started: time.Now(), shortcuts: map[ShortcutPurpose]string{},
	}
	defer func() {
		if value := recover(); value != nil {
			d.writeCrashReport("crash", fmt.Sprintf("%v\n%s", value, debug.Stack()))
			panic(value)
		}
	}()

	d.run()
}
