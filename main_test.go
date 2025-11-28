package main

import (
	"net"
	"testing"
)

// This test verifies that a link-local IPv6 address (fe80::/10)
// does NOT pass the selection logic used in getLocalIpv6.
func TestFilterLinkLocalIPv6(t *testing.T) {
	ip := net.ParseIP("fe80::8f27:6eda:2c95:9261")
	if ip == nil {
		t.Fatalf("failed to parse IP")
	}

	// Mirror the predicate used in getLocalIpv6
	allowed := ip.To4() == nil && ip.IsGlobalUnicast() && !ip.IsPrivate()

	if allowed {
		t.Fatalf("link-local IPv6 should be filtered out, but predicate allowed it")
	}
}

// Also verify that a typical global IPv6 (2001::/16 range) passes.
func TestAllowGlobalUnicastIPv6(t *testing.T) {
	ip := net.ParseIP("2001:db8::1")
	if ip == nil {
		t.Fatalf("failed to parse IP")
	}

	allowed := ip.To4() == nil && ip.IsGlobalUnicast() && !ip.IsPrivate()

	if !allowed {
		t.Fatalf("global unicast IPv6 should be allowed, but predicate filtered it")
	}
}
