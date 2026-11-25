package main

import (
	"reflect"
	"testing"
)

func TestDeepLinksQueuedBeforeMainReady(t *testing.T) {
	d := &Desktop{}
	first := deepLinkScheme + "://first"
	second := deepLinkScheme + "://second"
	d.receiveLinks([]string{"--flag", "https://example.com", first, first})
	d.receiveLinks([]string{first, second, "invalid"})
	if got := d.MainReady(); !reflect.DeepEqual(got, []string{first, second}) {
		t.Fatalf("queued links = %v", got)
	}
	if got := d.MainReady(); len(got) != 0 {
		t.Fatalf("queue was not drained: %v", got)
	}
}
