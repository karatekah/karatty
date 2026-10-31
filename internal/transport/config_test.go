package transport

import "testing"

func TestHTTPTransportPolicy(t *testing.T) {
	for _, addr := range []string{"0.0.0.0:8080", ":8080", "example.com:8080", "localhost:8080"} {
		if ValidateHTTP(addr, "", "", true) == nil {
			t.Fatalf("cleartext accepted on %s", addr)
		}
	}
	if ValidateHTTP("127.0.0.1:8080", "", "", false) == nil {
		t.Fatal("implicit cleartext")
	}
	if ValidateHTTP("127.0.0.1:8080", "", "", true) != nil {
		t.Fatal("explicit loopback dev rejected")
	}
}
