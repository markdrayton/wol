package main

import (
	"net/netip"
	"testing"
)

func TestBroadcastAddr(t *testing.T) {
	for _, tc := range []struct {
		prefix   netip.Prefix
		wantAddr netip.Addr
		wantOk   bool
	}{
		{
			prefix: netip.MustParsePrefix("::1/128"),
		},
		{
			prefix: netip.MustParsePrefix("fd72:638d:4b5d::/48"),
		},
		{
			prefix: netip.MustParsePrefix("::ffff:c0:a8:0:1/120"),
		},
		{
			prefix:   netip.MustParsePrefix("10.0.0.1/8"),
			wantAddr: netip.MustParseAddr("10.255.255.255"),
			wantOk:   true,
		},
		{
			prefix:   netip.MustParsePrefix("192.168.0.1/24"),
			wantAddr: netip.MustParseAddr("192.168.0.255"),
			wantOk:   true,
		},
	} {
		got, gotOk := broadcastAddr(tc.prefix)
		if got != tc.wantAddr || gotOk != tc.wantOk {
			t.Errorf("broadcastAddr(%v) = %v, %t; want %v, %t", tc.prefix, got, gotOk, tc.wantAddr, tc.wantOk)
		}
	}
}
