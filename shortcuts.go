package main

import (
	"errors"
	"fmt"
	"strings"
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
