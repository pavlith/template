package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"time"

	findfont "github.com/flopp/go-findfont"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
	"golang.org/x/image/font/sfnt"
)

func (d *Desktop) EnableAutostart() error            { return d.app.Autostart.Enable() }
func (d *Desktop) DisableAutostart() error           { return d.app.Autostart.Disable() }
func (d *Desktop) IsAutostartEnabled() (bool, error) { return d.app.Autostart.IsEnabled() }

func (d *Desktop) Notify(title, body string) (bool, error) {
	granted, err := d.notifier.CheckNotificationAuthorization()
	if err != nil {
		return false, err
	}
	if !granted {
		granted, err = d.notifier.RequestNotificationAuthorization()
		if err != nil || !granted {
			return granted, err
		}
	}
	return true, d.notifier.SendNotification(notifications.NotificationOptions{ID: fmt.Sprint(time.Now().UnixNano()), Title: title, Body: body})
}

func (d *Desktop) Greet(name string) string {
	return fmt.Sprintf("Hello, %s! You've been greeted from Go!", name)
}

// RunShellDemo executes a fixed command with no arguments supplied by the webview.
func (d *Desktop) RunShellDemo(ctx context.Context) (string, error) {
	const message = "Hello from the Wails shell demo"
	cmd := exec.CommandContext(ctx, "sh", "-c", "echo "+message)
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd.exe", "/C", "echo "+message)
	}
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func (d *Desktop) ListSystemFonts() []string {
	families := map[string]struct{}{}
	for _, path := range findfont.List() {
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		collection, err := sfnt.ParseCollectionReaderAt(file)
		var buffer sfnt.Buffer
		if err == nil {
			for i := 0; i < collection.NumFonts(); i++ {
				font, err := collection.Font(i)
				if err != nil {
					continue
				}
				name, err := font.Name(&buffer, sfnt.NameIDTypographicFamily)
				if err != nil {
					name, err = font.Name(&buffer, sfnt.NameIDFamily)
				}
				if err == nil && name != "" {
					families[name] = struct{}{}
				}
			}
		}
		file.Close()
	}
	result := []string{}
	for name := range families {
		result = append(result, name)
	}
	slices.SortFunc(result, func(a, b string) int { return strings.Compare(strings.ToLower(a), strings.ToLower(b)) })
	return result
}
