package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadAndSizeLimit(t *testing.T) {
	s := New(t.TempDir())
	defaults := func() map[string]int { return map[string]int{"default": 1} }
	if got, err := Load(s, "state.json", defaults); err != nil || got["default"] != 1 {
		t.Fatalf("defaults: %v, %v", got, err)
	}
	if err := Save(s, "state.json", map[string]int{"value": 2}, 0); err != nil {
		t.Fatal(err)
	}
	if err := Save(s, "state.json", map[string]int{"value": 3}, 1); err == nil {
		t.Fatal("oversized write succeeded")
	}
	if got, err := Load(s, "state.json", defaults); err != nil || got["value"] != 2 || got["default"] != 1 {
		t.Fatalf("stored value: %v, %v", got, err)
	}
	if files, err := filepath.Glob(filepath.Join(s.Dir(), ".tmp-*")); err != nil || len(files) != 0 {
		t.Fatalf("temporary files: %v, %v", files, err)
	}
}

func TestCorruptJSONBackup(t *testing.T) {
	s := New(t.TempDir())
	if err := os.WriteFile(filepath.Join(s.Dir(), "state.json"), []byte("!"), 0600); err != nil {
		t.Fatal(err)
	}
	if got, err := Load(s, "state.json", func() int { return 42 }); err != nil || got != 42 {
		t.Fatalf("recovery: %v, %v", got, err)
	}
	files, err := filepath.Glob(filepath.Join(s.Dir(), "state.json.corrupt-*"))
	if err != nil || len(files) != 1 {
		t.Fatalf("corrupt backups: %v, %v", files, err)
	}
}

func TestValidateFilename(t *testing.T) {
	for _, name := range []string{"", "../report", "a/b", "a\\b", ".hidden"} {
		if ValidateFilename(name) == nil {
			t.Errorf("accepted invalid name %q", name)
		}
	}
	if err := ValidateFilename("report-1"); err != nil {
		t.Fatal(err)
	}
}
