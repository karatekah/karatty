package main

import (
	"errors"
	"net"
)

func validatePublicMode(addr, cert, key, sshKey string, publicSignup, publicNetwork, dev bool) error {
	if publicNetwork && !publicSignup {
		return errors.New("public-network requires the encrypted public-signup mode")
	}
	if !publicSignup {
		return nil
	}
	if sshKey != "" {
		return errors.New("the encrypted public alpha cannot run a remote SSH TUI")
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	ip := net.ParseIP(host)
	if !publicNetwork && (ip == nil || !ip.IsLoopback()) {
		return errors.New("public listeners require the explicit public-network option")
	}
	if publicNetwork && (dev || cert == "" || key == "") {
		return errors.New("public-network requires a TLS certificate and key without dev mode")
	}
	return nil
}
