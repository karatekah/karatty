package main

import "testing"

func TestPublicNetworkRequiresEncryptedModeAndTLS(t *testing.T) {
	for _, tc := range []struct {
		name, addr, cert, key, ssh string
		public, network, dev       bool
		valid                      bool
	}{
		{"local experiment", "127.0.0.1:8080", "", "", "", true, false, true, true},
		{"public encrypted TLS", "0.0.0.0:443", "cert", "key", "", true, true, false, true},
		{"implicit public", "0.0.0.0:443", "cert", "key", "", true, false, false, false},
		{"legacy public option", "0.0.0.0:443", "cert", "key", "", false, true, false, false},
		{"missing TLS", "0.0.0.0:443", "", "", "", true, true, false, false},
		{"dev public", "0.0.0.0:443", "cert", "key", "", true, true, true, false},
		{"remote TUI", "0.0.0.0:443", "cert", "key", "ssh-key", true, true, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePublicMode(tc.addr, tc.cert, tc.key, tc.ssh, tc.public, tc.network, tc.dev)
			if (err == nil) != tc.valid {
				t.Fatalf("unexpected configuration result: %v", err)
			}
		})
	}
}
