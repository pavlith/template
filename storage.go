package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	appID                    = "com.pavlith.template"
	defaultQuickPaneShortcut = "CmdOrCtrl+Shift+."
	maxRecentItems           = 20
	maxThemeBytes            = 2 << 20
	maxRecoveryBytes         = 10 << 20
)

type ThemeProfile struct {
	PresetID   string  `json:"presetId"`
	Customized bool    `json:"customized"`
	Accent     string  `json:"accent"`
	Background string  `json:"background"`
	Foreground string  `json:"foreground"`
	Contrast   float64 `json:"contrast"`
}

type ImportedTheme struct {
	ID         string             `json:"id"`
	Name       string             `json:"name"`
	Mode       string             `json:"mode"`
	Accent     string             `json:"accent"`
	Background string             `json:"background"`
	Foreground string             `json:"foreground"`
	Contrast   float64            `json:"contrast"`
	Overrides  *map[string]string `json:"overrides"`
}

type AppPreferences struct {
	Theme             string             `json:"theme"`
	LightProfile      ThemeProfile       `json:"lightProfile"`
	DarkProfile       ThemeProfile       `json:"darkProfile"`
	ImportedThemes    []ImportedTheme    `json:"importedThemes"`
	FontFamily        *string            `json:"fontFamily"`
	FontSize          float64            `json:"fontSize"`
	ReducedMotion     string             `json:"reducedMotion"`
	PointerCursors    bool               `json:"pointerCursors"`
	WindowEffects     bool               `json:"windowEffects"`
	Language          *string            `json:"language"`
	GlobalShortcut    *string            `json:"globalShortcut"`
	QuickPaneShortcut *string            `json:"quickPaneShortcut"`
	CommandShortcuts  map[string]*string `json:"commandShortcuts"`
}

type PersistedAppState struct {
	LeftSidebarVisible  bool     `json:"leftSidebarVisible"`
	RightSidebarVisible bool     `json:"rightSidebarVisible"`
	SquareCorners       bool     `json:"squareCorners"`
	LastQuickPaneEntry  *string  `json:"lastQuickPaneEntry"`
	RecentItems         []string `json:"recentItems"`
	OnboardingCompleted bool     `json:"onboardingCompleted"`
}

func defaultPreferences() AppPreferences {
	shortcut := defaultQuickPaneShortcut
	return AppPreferences{
		Theme:             "system",
		LightProfile:      ThemeProfile{PresetID: "default-light", Accent: "#171717", Background: "#ffffff", Foreground: "#0a0a0a", Contrast: 50},
		DarkProfile:       ThemeProfile{PresetID: "default-dark", Accent: "#fafafa", Background: "#0a0a0a", Foreground: "#fafafa", Contrast: 50},
		ImportedThemes:    []ImportedTheme{},
		FontSize:          16,
		ReducedMotion:     "system",
		QuickPaneShortcut: &shortcut,
		CommandShortcuts:  map[string]*string{},
	}
}

func defaultAppState() PersistedAppState {
	return PersistedAppState{LeftSidebarVisible: true, RightSidebarVisible: true, RecentItems: []string{}}
}

var hexColor = regexp.MustCompile(`^#(?:[[:xdigit:]]{3}|[[:xdigit:]]{6})$`)

func validateThemeColors(name, accent, background, foreground string, contrast float64) error {
	if !hexColor.MatchString(accent) || !hexColor.MatchString(background) || !hexColor.MatchString(foreground) {
		return fmt.Errorf("%s colors must be hex colors", name)
	}
	if contrast < 0 || contrast > 100 {
		return fmt.Errorf("%s contrast must be between 0 and 100", name)
	}
	return nil
}

func (p AppPreferences) validate() error {
	switch p.Theme {
	case "light", "dark", "system":
	default:
		return errors.New("theme must be light, dark or system")
	}
	for _, profile := range []ThemeProfile{p.LightProfile, p.DarkProfile} {
		if profile.PresetID == "" {
			return errors.New("preset ID must not be empty")
		}
		if err := validateThemeColors(profile.PresetID, profile.Accent, profile.Background, profile.Foreground, profile.Contrast); err != nil {
			return err
		}
	}
	for _, theme := range p.ImportedThemes {
		if theme.ID == "" || theme.Name == "" || (theme.Mode != "light" && theme.Mode != "dark") {
			return fmt.Errorf("invalid imported theme %q", theme.ID)
		}
		if err := validateThemeColors(theme.ID, theme.Accent, theme.Background, theme.Foreground, theme.Contrast); err != nil {
			return err
		}
	}
	if p.FontSize < 12 || p.FontSize > 20 {
		return errors.New("fontSize must be between 12 and 20")
	}
	switch p.ReducedMotion {
	case "system", "on", "off":
	default:
		return errors.New("reducedMotion must be system, on or off")
	}
	for id, combo := range p.CommandShortcuts {
		if id == "" || (combo != nil && *combo == "") {
			return errors.New("commandShortcuts requires nonempty IDs and a combo or null")
		}
	}
	return nil
}

// Use the template application ID and keep the persisted JSON schema stable.
func dataDirectory() (string, error) {
	var dir string
	var err error
	switch runtime.GOOS {
	case "linux":
		dir = os.Getenv("XDG_DATA_HOME")
		if !filepath.IsAbs(dir) {
			dir, err = os.UserHomeDir()
			dir = filepath.Join(dir, ".local", "share")
		}
	default:
		dir, err = os.UserConfigDir()
	}
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appID), nil
}

type storage struct {
	dir string
	mu  sync.Mutex
}

func loadJSON[T any](s *storage, name string, defaults func() T) (T, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.dir, name)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return defaults(), nil
	}
	if err != nil {
		return defaults(), err
	}
	value := defaults()
	if err := json.Unmarshal(data, &value); err != nil {
		backup := fmt.Sprintf("%s.corrupt-%d", path, time.Now().UnixNano())
		if renameErr := os.Rename(path, backup); renameErr != nil {
			return defaults(), fmt.Errorf("backup %s: %w", name, renameErr)
		}
		log.Printf("%s: %v (backed up to %s)", name, err, backup)
		return defaults(), nil
	}
	return value, nil
}

func saveJSON(s *storage, name string, value any, maxBytes int) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if maxBytes > 0 && len(data) > maxBytes {
		return errors.New("JSON exceeds the file size limit")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	path := filepath.Join(s.dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	_, writeErr := tmp.Write(append(data, '\n'))
	syncErr := tmp.Sync()
	if err := errors.Join(writeErr, syncErr, tmp.Close()); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func validateFilename(name string) error {
	if name == "" || len(name) > 200 || strings.HasPrefix(name, ".") ||
		strings.Contains(name, "..") || strings.ContainsAny(name, "/\\:\x00") {
		return errors.New("invalid filename")
	}
	return nil
}

func readLimited(path string, limit int64) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, errors.New("expected a regular file within the size limit")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("file exceeds the size limit")
	}
	return data, nil
}

func cleanupReports(dir, extension string, olderThan time.Duration) (int, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var removeErr error
	removed := 0
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != extension || !entry.Type().IsRegular() {
			continue
		}
		if olderThan > 0 {
			info, err := entry.Info()
			if err != nil || time.Since(info.ModTime()) <= olderThan {
				continue
			}
		}
		err := os.Remove(filepath.Join(dir, entry.Name()))
		if err == nil {
			removed++
		} else if !errors.Is(err, os.ErrNotExist) {
			removeErr = errors.Join(removeErr, err)
		}
	}
	return removed, removeErr
}
