package api

import (
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestReactAppBridgeReadinessTimeoutReturnsErrorAndClearsProxy(t *testing.T) {
	oldTimeout := reactAppReadyTimeout
	oldPoll := reactAppReadyPollInterval
	reactAppReadyTimeout = 30 * time.Millisecond
	reactAppReadyPollInterval = 5 * time.Millisecond
	t.Cleanup(func() {
		reactAppReadyTimeout = oldTimeout
		reactAppReadyPollInterval = oldPoll
	})

	port, err := freePort()
	if err != nil {
		t.Fatalf("freePort() error: %v", err)
	}
	u, err := url.Parse("http://127.0.0.1:1/")
	if err != nil {
		t.Fatalf("url.Parse() error: %v", err)
	}
	b := newReactAppBridge()
	b.status = "running"
	b.base = u
	b.proxy = httputil.NewSingleHostReverseProxy(u)

	err = b.waitUntilReady(port, nil)
	if err == nil {
		t.Fatalf("waitUntilReady() error = nil, want timeout error")
	}
	if !strings.Contains(err.Error(), "did not become ready") {
		t.Fatalf("waitUntilReady() error = %q, want readiness timeout", err.Error())
	}
	if b.status != "stopped" {
		t.Fatalf("status = %q, want stopped", b.status)
	}
	if b.proxy != nil || b.base != nil {
		t.Fatalf("proxy/base not cleared after timeout: proxy=%v base=%v", b.proxy, b.base)
	}
	joined := strings.Join(b.logs, "\n")
	if !strings.Contains(joined, "[readiness timeout]") {
		t.Fatalf("logs = %q, want readiness timeout entry", joined)
	}
}
