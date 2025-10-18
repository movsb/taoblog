package geoip

import (
	"net/netip"
	"testing"
	"time"
)

func TestQuery(t *testing.T) {
	if IsInChina(netip.MustParseAddr(`185.201.226.166`)) {
		t.Fatal(`not china ip`)
	}
	if !IsInChina(netip.MustParseAddr(`183.9.2.91`)) {
		t.Fatal(`china ip`)
	}
}

func TestInvalidate(t *testing.T) {
	next := time.Date(2026, 7, 3, 0, 0, 0, 0, time.Local)
	if time.Now().After(next) {
		t.Fatal(`应该更新 GeoIP 数据库了。`)
	}
}
