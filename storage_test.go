package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadLegacyPreferences(t *testing.T) {
	store := &storage{dir: t.TempDir()}
	path := filepath.Join(store.dir, "preferences.json")
	for _, tc := range []struct {
		json              string
		explicitlyCleared bool
	}{
		{`{"theme":"dark"}`, false},
		{`{"theme":"dark","quickPaneShortcut":null}`, true},
	} {
		if err := os.WriteFile(path, []byte(tc.json), 0600); err != nil {
			t.Fatal(err)
		}
		prefs, err := loadJSON(store, "preferences.json", defaultPreferences)
		if err != nil {
			t.Fatal(err)
		}
		if prefs.Theme != "dark" || prefs.FontSize != 16 {
			t.Fatalf("unexpected preferences: %+v", prefs)
		}
		if tc.explicitlyCleared != (prefs.QuickPaneShortcut == nil) {
			t.Fatalf("unexpected shortcut for %s: %v", tc.json, prefs.QuickPaneShortcut)
		}
	}
}

func TestCorruptJSONMovedAside(t *testing.T) {
	store := &storage{dir: t.TempDir()}
	path := filepath.Join(store.dir, "state.json")
	if err := os.WriteFile(path, []byte("!"), 0600); err != nil {
		t.Fatal(err)
	}
	state, err := loadJSON(store, "state.json", defaultAppState)
	if err != nil || !state.LeftSidebarVisible {
		t.Fatalf("state %+v, error %v", state, err)
	}
	files, err := filepath.Glob(path + ".corrupt-*")
	if err != nil || len(files) != 1 {
		t.Fatalf("backups %v, error %v", files, err)
	}
}
