package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/pavlith/template/internal/storage"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const preferencesFile = "preferences.json"

func (d *Desktop) LoadPreferences() (AppPreferences, error) {
	return storage.Load(d.store, preferencesFile, defaultPreferences)
}

func (d *Desktop) SavePreferences(prefs AppPreferences) error {
	if err := prefs.validate(); err != nil {
		return err
	}
	return storage.Save(d.store, preferencesFile, prefs, 0)
}

func (d *Desktop) OpenPreferencesFile() error {
	path := filepath.Join(d.store.Dir(), preferencesFile)
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return d.app.Browser.OpenFile(path)
}

func (d *Desktop) LoadAppState() (PersistedAppState, error) {
	return storage.Load(d.store, "state.json", defaultAppState)
}

func (d *Desktop) SaveAppState(state PersistedAppState) error {
	if len(state.RecentItems) > maxRecentItems {
		state.RecentItems = state.RecentItems[:maxRecentItems]
	}
	return storage.Save(d.store, "state.json", state, 0)
}

type PickedTheme struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

func (d *Desktop) PickThemeFile(ctx context.Context, title, filter string) (*PickedTheme, error) {
	dialog := d.app.Dialog.OpenFile().CanChooseFiles(true).SetTitle(title).AddFilter(filter, "*.json;*.jsonc")
	if window, ok := ctx.Value(application.WindowKey).(application.Window); ok {
		dialog.AttachToWindow(window)
	}
	path, err := dialog.PromptForSingleSelection()
	if err != nil || path == "" {
		return nil, err
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json", ".jsonc":
	default:
		return nil, errors.New("theme must be a JSON or JSONC file")
	}
	data, err := storage.ReadLimited(path, maxThemeBytes)
	if err != nil {
		return nil, err
	}
	return &PickedTheme{Name: filepath.Base(path), Text: string(data)}, nil
}
