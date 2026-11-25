package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"github.com/pavlith/template/internal/storage"
)

func (d *Desktop) SaveEmergencyData(filename string, data any) error {
	if err := storage.ValidateFilename(filename); err != nil {
		return err
	}
	return storage.Save(d.store, filepath.Join("recovery", filename+".json"), data, maxRecoveryBytes)
}

func (d *Desktop) CleanupOldRecoveryFiles() (int, error) {
	return storage.CleanupReports(filepath.Join(d.store.Dir(), "recovery"), ".json", 7*24*time.Hour)
}

type CrashReportSummary struct {
	Filename      string `json:"filename"`
	TimestampSecs int64  `json:"timestampSecs"`
	SecondsAgo    int64  `json:"secondsAgo"`
}

func (d *Desktop) crashDir() string { return filepath.Join(d.store.Dir(), "crash-reports") }

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
	if err := storage.ValidateFilename(name); err != nil {
		return "", err
	}
	if filepath.Ext(name) != ".log" {
		return "", errors.New("expected a crash report")
	}
	data, err := storage.ReadLimited(filepath.Join(d.crashDir(), name), maxRecoveryBytes)
	return string(data), err
}

func (d *Desktop) ClearCrashReports() (int, error) {
	return storage.CleanupReports(d.crashDir(), ".log", 0)
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
