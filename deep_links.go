package main

func (d *Desktop) receiveLinks(args []string) {
	if links := d.links.Receive(deepLinkScheme, args); len(links) > 0 {
		d.main.EmitEvent("app:deep-link-received", links)
	}
}

func (d *Desktop) MainReady() []string { return d.links.Ready() }
