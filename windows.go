package main

import (
	"log"
	"math"
	"runtime"

	"github.com/pavlith/template/internal/storage"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

func (d *Desktop) createWindows() {
	state, err := storage.Load(d.store, "window.json", func() windowState { return windowState{} })
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
	if err := storage.Save(d.store, "window.json", state, 0); err != nil {
		log.Print(err)
	}
}

func (d *Desktop) showMain() {
	if !d.links.IsReady() {
		return
	}
	d.main.UnMinimise()
	d.main.Show()
	d.main.Focus()
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
