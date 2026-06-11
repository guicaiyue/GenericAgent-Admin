package main

import (
	"net/http"
	"testing"
)

func TestNewHTTPServerSetsTimeouts(t *testing.T) {
	handler := http.NewServeMux()
	server := newHTTPServer("127.0.0.1:0", handler)

	if server.Addr != "127.0.0.1:0" {
		t.Fatalf("Addr = %q, want %q", server.Addr, "127.0.0.1:0")
	}
	if server.Handler != handler {
		t.Fatal("Handler was not preserved")
	}
	if server.ReadHeaderTimeout != adminReadHeaderTimeout {
		t.Fatalf("ReadHeaderTimeout = %v, want %v", server.ReadHeaderTimeout, adminReadHeaderTimeout)
	}
	if server.IdleTimeout != adminIdleTimeout {
		t.Fatalf("IdleTimeout = %v, want %v", server.IdleTimeout, adminIdleTimeout)
	}
	if server.ReadHeaderTimeout <= 0 {
		t.Fatal("ReadHeaderTimeout must be positive")
	}
	if server.IdleTimeout <= server.ReadHeaderTimeout {
		t.Fatalf("IdleTimeout = %v must exceed ReadHeaderTimeout = %v", server.IdleTimeout, server.ReadHeaderTimeout)
	}
}
