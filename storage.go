package main

import (
	"errors"
	"fmt"
	"regexp"
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
