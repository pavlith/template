package main

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/pavlith/template/internal/deeplink"
	"github.com/pavlith/template/internal/storage"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
)

type Desktop struct {
	app        *application.App
	main       *application.WebviewWindow
	pane       *application.WebviewWindow
	store      *storage.Store
	notifier   *notifications.NotificationService
	version    string
	started    time.Time
	relaunch   atomic.Bool
	shortcutMu sync.Mutex
	shortcuts  map[ShortcutPurpose]string
	links      deeplink.Queue
}
