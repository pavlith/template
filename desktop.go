package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	findfont "github.com/flopp/go-findfont"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/endpoint"
	"golang.org/x/image/font/sfnt"
)

type ShortcutPurpose string

const (
	FocusMain ShortcutPurpose = "focusMain"
	QuickPane ShortcutPurpose = "quickPane"
)

func normalizeAccelerator(accel string) string {
	parts := strings.Split(accel, "+")
	for i, part := range parts {
		switch strings.ToLower(strings.TrimSpace(part)) {
		case "commandorcontrol", "commandorctrl", "cmdorctrl":
			parts[i] = "CmdOrCtrl"
		case "control":
			parts[i] = "Ctrl"
		case "meta", "super":
			parts[i] = "Super"
		default:
			parts[i] = strings.TrimSpace(part)
		}
	}
	return strings.Join(parts, "+")
}

func (d *Desktop) RegisterGlobalShortcut(purpose ShortcutPurpose, accel string) error {
	var callback func()
	switch purpose {
	case FocusMain:
		callback = d.showMain
	case QuickPane:
		callback = d.ToggleQuickPane
	default:
		return fmt.Errorf("invalid shortcut purpose %q", purpose)
	}

	accel = normalizeAccelerator(accel)
	if len(accel) > 128 || !strings.Contains(accel, "+") {
		return errors.New("shortcut requires a modifier and a key")
	}

	d.shortcutMu.Lock()
	defer d.shortcutMu.Unlock()
	old := d.shortcuts[purpose]
	if strings.EqualFold(old, accel) {
		return nil
	}
	if err := d.app.GlobalShortcut.Register(accel, callback); err != nil {
		return err
	}
	if old != "" {
		if err := d.app.GlobalShortcut.Unregister(old); err != nil {
			return errors.Join(err, d.app.GlobalShortcut.Unregister(accel))
		}
	}
	d.shortcuts[purpose] = accel
	return nil
}

func (d *Desktop) UnregisterGlobalShortcut(purpose ShortcutPurpose) error {
	d.shortcutMu.Lock()
	defer d.shortcutMu.Unlock()
	if accel := d.shortcuts[purpose]; accel != "" {
		if err := d.app.GlobalShortcut.Unregister(accel); err != nil {
			return err
		}
		delete(d.shortcuts, purpose)
	}
	return nil
}

func (d *Desktop) receiveLinks(args []string) {
	links := []string{}
	for _, arg := range args {
		u, err := url.Parse(arg)
		if err == nil && strings.EqualFold(u.Scheme, deepLinkScheme) && !slices.Contains(links, arg) {
			links = append(links, arg)
		}
	}
	if len(links) == 0 {
		return
	}
	d.linkMu.Lock()
	if !d.mainReady {
		for _, link := range links {
			if !slices.Contains(d.pendingLinks, link) {
				d.pendingLinks = append(d.pendingLinks, link)
			}
		}
		d.linkMu.Unlock()
		return
	}
	d.linkMu.Unlock()
	d.main.EmitEvent("app:deep-link-received", links)
}

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

// Configure a signed update feed at build time via linker flags.
var updateFeedURL string
var updatePublicKey string

func (d *Desktop) initUpdater() {
	if updateFeedURL == "" {
		return
	}
	key, err := base64.StdEncoding.DecodeString(updatePublicKey)
	if err != nil || len(key) == 0 {
		log.Print("Updater requires a base64-encoded public key")
		return
	}
	provider, err := endpoint.New(endpoint.Config{URL: updateFeedURL})
	if err == nil {
		err = d.app.Updater.Init(updater.Config{CurrentVersion: d.version, Providers: []updater.Provider{provider}, PublicKey: key, Window: updater.WindowNone})
	}
	if err != nil {
		log.Printf("initialize updater: %v", err)
	}
}

func (d *Desktop) UpdateConfigured() bool { return d.app.Updater.CurrentVersion() != "" }

func (d *Desktop) CheckForUpdate(ctx context.Context) (bool, error) {
	release, err := d.app.Updater.Check(ctx)
	return release != nil, err
}

func (d *Desktop) InstallUpdate(ctx context.Context) error {
	return d.app.Updater.DownloadAndInstall(ctx)
}
func (d *Desktop) RestartForUpdate(ctx context.Context) error { return d.app.Updater.Restart(ctx) }

const preferencesFile = "preferences.json"

func (d *Desktop) LoadPreferences() (AppPreferences, error) {
	return loadJSON(&d.store, preferencesFile, defaultPreferences)
}

func (d *Desktop) SavePreferences(prefs AppPreferences) error {
	if err := prefs.validate(); err != nil {
		return err
	}
	return saveJSON(&d.store, preferencesFile, prefs, 0)
}

func (d *Desktop) OpenPreferencesFile() error {
	path := filepath.Join(d.store.dir, preferencesFile)
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return d.app.Browser.OpenFile(path)
}

func (d *Desktop) LoadAppState() (PersistedAppState, error) {
	return loadJSON(&d.store, "state.json", defaultAppState)
}

func (d *Desktop) SaveAppState(state PersistedAppState) error {
	if len(state.RecentItems) > maxRecentItems {
		state.RecentItems = state.RecentItems[:maxRecentItems]
	}
	return saveJSON(&d.store, "state.json", state, 0)
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
	data, err := readLimited(path, maxThemeBytes)
	if err != nil {
		return nil, err
	}
	return &PickedTheme{Name: filepath.Base(path), Text: string(data)}, nil
}

func (d *Desktop) SaveEmergencyData(filename string, data any) error {
	if err := validateFilename(filename); err != nil {
		return err
	}
	return saveJSON(&d.store, filepath.Join("recovery", filename+".json"), data, maxRecoveryBytes)
}

func (d *Desktop) CleanupOldRecoveryFiles() (int, error) {
	return cleanupReports(filepath.Join(d.store.dir, "recovery"), ".json", 7*24*time.Hour)
}

type CrashReportSummary struct {
	Filename      string `json:"filename"`
	TimestampSecs int64  `json:"timestampSecs"`
	SecondsAgo    int64  `json:"secondsAgo"`
}

func (d *Desktop) crashDir() string { return filepath.Join(d.store.dir, "crash-reports") }

func (d *Desktop) writeCrashReport(prefix, message string) error {
	log.Print(message)
	dir := d.crashDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path := filepath.Join(dir, prefix+"-"+time.Now().UTC().Format("20060102T150405.000000000")+".log")
	return os.WriteFile(path, []byte(message+"\n"), 0600)
}

func (d *Desktop) LogFrontend(level, message string) {
	log.Printf("frontend %s: %s", level, message)
}

func (d *Desktop) LogFrontendError(message string, stack, componentStack *string) error {
	for _, trace := range []*string{stack, componentStack} {
		if trace != nil {
			message += "\n" + *trace
		}
	}
	return d.writeCrashReport("frontend-error", message)
}

func (d *Desktop) listCrashReports() ([]CrashReportSummary, error) {
	entries, err := os.ReadDir(d.crashDir())
	reports := []CrashReportSummary{}
	if errors.Is(err, os.ErrNotExist) {
		return reports, nil
	}
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".log" || !entry.Type().IsRegular() {
			continue
		}
		info, err := entry.Info()
		if err == nil {
			reports = append(reports, CrashReportSummary{entry.Name(), info.ModTime().Unix(), max(0, int64(time.Since(info.ModTime()).Seconds()))})
		}
	}
	slices.SortFunc(reports, func(a, b CrashReportSummary) int {
		if a.TimestampSecs > b.TimestampSecs {
			return -1
		}
		if a.TimestampSecs < b.TimestampSecs {
			return 1
		}
		return strings.Compare(b.Filename, a.Filename)
	})
	return reports, nil
}

func (d *Desktop) HasRecentCrash() (*CrashReportSummary, error) {
	reports, err := d.listCrashReports()
	if err != nil || len(reports) == 0 || reports[0].SecondsAgo > 300 {
		return nil, err
	}
	return &reports[0], nil
}

func (d *Desktop) ListCrashReports() ([]string, error) {
	reports, err := d.listCrashReports()
	names := []string{}
	for _, report := range reports {
		names = append(names, report.Filename)
	}
	return names, err
}

func (d *Desktop) GetCrashReport(name string) (string, error) {
	if err := validateFilename(name); err != nil {
		return "", err
	}
	if filepath.Ext(name) != ".log" {
		return "", errors.New("expected a crash report")
	}
	data, err := readLimited(filepath.Join(d.crashDir(), name), maxRecoveryBytes)
	return string(data), err
}

func (d *Desktop) ClearCrashReports() (int, error) {
	return cleanupReports(d.crashDir(), ".log", 0)
}

type SystemInfo struct {
	AppName      string `json:"appName"`
	AppVersion   string `json:"appVersion"`
	WailsVersion string `json:"wailsVersion"`
	OSName       string `json:"osName"`
	OSArch       string `json:"osArch"`
	OSVersion    string `json:"osVersion"`
}

func (d *Desktop) GetSystemInfo() SystemInfo {
	version := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == "github.com/wailsapp/wails/v3" {
				version = dep.Version
				break
			}
		}
	}
	osVersion := ""
	if info := d.app.Env.Info().OSInfo; info != nil {
		osVersion = info.Version
	}
	return SystemInfo{appName, d.version, version, runtime.GOOS, runtime.GOARCH, osVersion}
}

type DiagnosticsReport struct {
	SystemInfo
	Settings           map[string]any `json:"settings"`
	RecentCrashReports []string       `json:"recentCrashReports"`
	MemoryUsageBytes   uint64         `json:"memoryUsageBytes"`
	UptimeSecs         int64          `json:"uptimeSecs"`
}

func (d *Desktop) CollectDiagnostics() (DiagnosticsReport, error) {
	prefs, err := d.LoadPreferences()
	if err != nil {
		return DiagnosticsReport{}, err
	}
	reports, err := d.ListCrashReports()
	if err != nil {
		return DiagnosticsReport{}, err
	}
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return DiagnosticsReport{
		SystemInfo: d.GetSystemInfo(),
		Settings: map[string]any{
			"theme": prefs.Theme, "language": prefs.Language, "fontSize": prefs.FontSize,
			"reducedMotion": prefs.ReducedMotion, "pointerCursors": prefs.PointerCursors, "windowEffects": prefs.WindowEffects,
		},
		RecentCrashReports: reports[:min(10, len(reports))],
		MemoryUsageBytes:   mem.Alloc, UptimeSecs: int64(time.Since(d.started).Seconds()),
	}, nil
}
