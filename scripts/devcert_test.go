package main

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"
)

func TestDevelopmentCertificateLAN(t *testing.T) {
	dir := t.TempDir()
	if err := run(dir, "192.168.18.6"); err != nil {
		t.Fatal(err)
	}
	caPEM, err := os.ReadFile(filepath.Join(dir, "ca.crt"))
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(caPEM) {
		t.Fatal("invalid CA")
	}
	pair, err := tls.LoadX509KeyPair(filepath.Join(dir, "server.crt"), filepath.Join(dir, "server.key"))
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(pair.Certificate[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, host := range []string{"localhost", "127.0.0.1", "::1", "192.168.18.6"} {
		if _, err := cert.Verify(x509.VerifyOptions{Roots: roots, DNSName: host}); err != nil {
			t.Fatalf("%s: %v", host, err)
		}
	}
	if _, err := cert.Verify(x509.VerifyOptions{Roots: roots, DNSName: "192.168.18.7"}); err == nil {
		t.Fatal("accepted an IP absent from the certificate")
	}
	if err := run(dir, "192.168.18.7"); err == nil {
		t.Fatal("overwrote existing certificates")
	}
}

func TestDevelopmentCertificateRejectsInvalidIP(t *testing.T) {
	for _, address := range []string{"invalid", "0.0.0.0", "::", "224.0.0.1"} {
		if err := run(t.TempDir(), address); err == nil {
			t.Fatalf("accepted %q", address)
		}
	}
}
