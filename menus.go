package main

import (
	"context"
	"runtime"
	"strconv"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type NativeMenuItem struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Role        string `json:"role"`
	Accelerator string `json:"accelerator"`
	Disabled    bool   `json:"disabled"`
}

type NativeSubmenu struct {
	Label string           `json:"label"`
	Items []NativeMenuItem `json:"items"`
}

var menuRoles = map[string]application.Role{
	"About": application.About,
	"Undo":  application.Undo, "Redo": application.Redo, "Cut": application.Cut,
	"Copy": application.Copy, "Paste": application.Paste, "SelectAll": application.SelectAll,
	"Hide": application.Hide, "HideOthers": application.HideOthers, "ShowAll": application.ShowAll,
	"Minimize": application.Minimise, "Maximize": application.Zoom,
}

func addMenuItems(menu *application.Menu, items []NativeMenuItem, onClick func(string)) {
	for _, item := range items {
		switch {
		case item.Role == "Separator":
			menu.AddSeparator()
		case item.Role != "":
			if role, ok := menuRoles[item.Role]; ok {
				menu.AddRole(role)
			}
		default:
			entry := menu.Add(item.Label).SetEnabled(!item.Disabled)
			if item.Accelerator != "" {
				entry.SetAccelerator(normalizeAccelerator(item.Accelerator))
			}
			entry.OnClick(func(*application.Context) { onClick(item.ID) })
		}
	}
}

func (d *Desktop) SetAppMenu(submenus []NativeSubmenu) {
	if runtime.GOOS != "darwin" {
		return
	}
	menu := d.app.Menu.New()
	for _, sub := range submenus {
		addMenuItems(menu.AddSubmenu(sub.Label), sub.Items, func(id string) {
			d.main.EmitEvent("app:menu-command", id)
		})
	}
	d.app.Menu.Set(menu)
}

type ContextSelection struct {
	Request int    `json:"request"`
	ID      string `json:"id"`
}

func (d *Desktop) ShowContextMenu(ctx context.Context, request, x, y int, items []NativeMenuItem) {
	window, ok := ctx.Value(application.WindowKey).(application.Window)
	if !ok {
		return
	}
	application.InvokeSync(func() {
		name := "template-context-" + window.Name() + "-" + strconv.Itoa(request)
		menu := d.app.ContextMenu.New()
		addMenuItems(menu.Menu, items, func(id string) {
			window.EmitEvent("app:context-selection", ContextSelection{request, id})
		})
		d.app.ContextMenu.Add(name, menu)
		defer d.app.ContextMenu.Remove(name)
		window.OpenContextMenu(&application.ContextMenuData{Id: name, X: x, Y: y})
	})
}
