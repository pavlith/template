package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
)

//go:embed build/appicon.png
var appIcon []byte

// package.json also supplies the frontend and distribution version.
//
//go:embed package.json
var packageJSON []byte

const appName = "Wails App"
const deepLinkScheme = "tauri-app"

type Desktop struct {
	app          *application.App
	main         *application.WebviewWindow
	pane         *application.WebviewWindow
	store        storage
	notifier     *notifications.NotificationService
	version      string
	started      time.Time
	relaunch     atomic.Bool
	shortcutMu   sync.Mutex
	shortcuts    map[ShortcutPurpose]string
	linkMu       sync.Mutex
	pendingLinks []string
	mainReady    bool
}

func main() {
	dir, err := dataDirectory()
	if err != nil {
		log.Fatal(err)
	}
	var pkg struct{ Version string }
	if err := json.Unmarshal(packageJSON, &pkg); err != nil {
		log.Fatal(err)
	}
	d := &Desktop{
		store: storage{dir: dir}, notifier: notifications.New(), version: pkg.Version,
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

func (d *Desktop) run() {
	var assets http.Handler
	if frontendAssets != nil {
		assets = application.AssetFileServerFS(frontendAssets)
	}
	d.app = application.New(application.Options{
		Name: appName, Description: "A Wails and Svelte desktop template", Icon: appIcon,
		Assets:   application.AssetOptions{Handler: assets},
		Linux:    application.LinuxOptions{ApplicationID: appID},
		Services: []application.Service{application.NewService(d)},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: appID,
			OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
				d.receiveLinks(data.Args)
				d.showMain()
			},
		},
		PanicHandler: func(p *application.PanicDetails) {
			d.writeCrashReport("crash", fmt.Sprintf("%v\n%s", p.Error, p.FullStackTrace))
		},
	})
	d.app.RegisterService(application.NewService(d.notifier))
	d.initUpdater()
	d.createWindows()

	d.app.Event.OnApplicationEvent(events.Common.ApplicationLaunchedWithUrl, func(event *application.ApplicationEvent) {
		d.receiveLinks([]string{event.Context().URL()})
		d.showMain()
	})
	if runtime.GOOS == "darwin" {
		d.app.Event.OnApplicationEvent(events.Mac.ApplicationShouldHandleReopen, func(*application.ApplicationEvent) { d.showMain() })
	}
	d.app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
		prefs, err := d.LoadPreferences()
		if err != nil {
			log.Print(err)
			return
		}
		for purpose, shortcut := range map[ShortcutPurpose]*string{FocusMain: prefs.GlobalShortcut, QuickPane: prefs.QuickPaneShortcut} {
			if shortcut != nil {
				if err := d.RegisterGlobalShortcut(purpose, *shortcut); err != nil {
					log.Printf("register %s shortcut: %v", purpose, err)
				}
			}
		}
	})

	tray := d.app.SystemTray.New().SetIcon(appIcon).OnClick(d.showMain)
	menu := d.app.Menu.New()
	menu.Add("Show").OnClick(func(*application.Context) { d.showMain() })
	menu.Add("Quit").OnClick(func(*application.Context) {
		d.showMain()
		d.main.EmitEvent("tray:quit-requested")
	})
	tray.SetMenu(menu)
	d.app.OnShutdown(d.saveWindowState)
	if err := d.app.Run(); err != nil {
		log.Print(err)
		return
	}
	if d.relaunch.Load() {
		executable, err := os.Executable()
		if err == nil {
			command := exec.Command(executable)
			command.Stdout, command.Stderr = os.Stdout, os.Stderr
			err = command.Start()
			if err == nil {
				_ = command.Process.Release()
			}
		}
		if err != nil {
			log.Printf("relaunch: %v", err)
		}
	}
}

func (d *Desktop) createWindows() {
	state, err := loadJSON(&d.store, "window.json", func() windowState { return windowState{} })
	if err != nil {
		log.Print(err)
	}
	main := application.WebviewWindowOptions{
		Name: "main", Title: appName, URL: "/", Width: 1200, Height: 800, MinWidth: 1000, MinHeight: 700,
		Hidden: true, Frameless: runtime.GOOS != "linux", EnableFileDrop: true,
		Windows: application.WindowsWindow{BackdropType: application.Auto},
		Mac:     application.MacWindow{Backdrop: application.MacBackdropTranslucent},
	}
	if runtime.GOOS == "windows" {
		main.BackgroundType = application.BackgroundTypeTranslucent
	}
	if state.Width >= main.MinWidth && state.Height >= main.MinHeight {
		main.Width, main.Height = state.Width, state.Height
		if runtime.GOOS != "linux" {
			main.InitialPosition, main.X, main.Y = application.WindowXY, state.X, state.Y
		}
		if state.Maximised {
			main.StartState = application.WindowStateMaximised
		}
	}
	d.main = d.app.Window.NewWithOptions(main)
	d.main.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		event.Cancel()
		d.main.EmitEvent("app:close-requested", runtime.GOOS == "darwin")
	})
	d.pane = d.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name: "quick-pane", Title: "Quick Pane", URL: "/quick-pane.html", Width: 640, Height: 84, DisableResize: true,
		Frameless: true, AlwaysOnTop: true, Hidden: true, BackgroundType: application.BackgroundTypeTransparent,
		Mac: application.MacWindow{
			Backdrop:           application.MacBackdropTransparent,
			CollectionBehavior: application.MacWindowCollectionBehaviorCanJoinAllSpaces | application.MacWindowCollectionBehaviorFullScreenAuxiliary,
		},
	})
	d.pane.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		event.Cancel()
		d.pane.Hide()
	})
	d.pane.OnWindowEvent(events.Common.WindowLostFocus, func(*application.WindowEvent) { d.pane.Hide() })
}

type windowState struct {
	X         int  `json:"x"`
	Y         int  `json:"y"`
	Width     int  `json:"width"`
	Height    int  `json:"height"`
	Maximised bool `json:"maximised"`
}

func (d *Desktop) saveWindowState() {
	bounds := d.main.Bounds()
	if bounds.Width < 1000 || bounds.Height < 700 {
		return
	}
	state := windowState{bounds.X, bounds.Y, bounds.Width, bounds.Height, d.main.IsMaximised()}
	if err := saveJSON(&d.store, "window.json", state, 0); err != nil {
		log.Print(err)
	}
}

func (d *Desktop) showMain() {
	d.linkMu.Lock()
	ready := d.mainReady
	d.linkMu.Unlock()
	if !ready {
		return
	}
	d.main.UnMinimise()
	d.main.Show()
	d.main.Focus()
}

func (d *Desktop) MainReady() []string {
	d.linkMu.Lock()
	defer d.linkMu.Unlock()
	d.mainReady = true
	links := d.pendingLinks
	d.pendingLinks = nil
	return links
}

func (d *Desktop) QuitApp() { d.app.Quit() }

func (d *Desktop) Relaunch() {
	d.relaunch.Store(true)
	d.app.Quit()
}

func (d *Desktop) ShowQuickPane() {
	if screen, err := d.main.GetScreen(); err == nil && screen != nil {
		area := screen.WorkArea
		d.pane.SetScreen(screen)
		d.pane.SetPosition(area.X+(area.Width-640)/2, area.Y+int(math.Round(float64(area.Height-84)*0.25)))
	}
	d.pane.Show()
	d.pane.Focus()
}

func (d *Desktop) DismissQuickPane() { d.pane.Hide() }

func (d *Desktop) ToggleQuickPane() {
	if d.pane.IsVisible() {
		d.DismissQuickPane()
		return
	}
	d.ShowQuickPane()
}
