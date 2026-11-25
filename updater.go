package main

import (
	"context"
	"encoding/base64"
	"log"

	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/endpoint"
)

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
