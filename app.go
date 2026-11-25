package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

func (d *Desktop) run() {
	// Queue command-line deep links before the frontend registers its listeners.
	d.receiveLinks(os.Args[1:])
	d.app = application.New(application.Options{
		Name: appName, Description: "A Wails and Svelte desktop template", Icon: appIcon,
		Assets:   application.AssetOptions{Handler: application.AssetFileServerFS(frontendAssets)},
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
		log.Fatal(err)
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
