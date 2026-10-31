package transport

import (
	"errors"
	"net"
)

func ValidateHTTP(addr, cert, key string, development bool) error {
	if cert != "" && key != "" {
		return nil
	}
	if cert != "" || key != "" {
		return errors.New("both TLS certificate and key are required")
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	ip := net.ParseIP(host)
	if !development || ip == nil || !ip.IsLoopback() {
		return errors.New("TLS required; --dev permits HTTP on loopback only")
	}
	return nil
}
