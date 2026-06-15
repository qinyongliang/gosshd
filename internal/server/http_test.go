package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunSHUsesRunDownloadFlow(t *testing.T) {
	app := NewApp(Config{PublicHost: "relay.example.com", AgentToken: "secret"})
	req := httptest.NewRequest(http.MethodGet, "http://relay.example.com/run.sh", nil)
	rec := httptest.NewRecorder()

	app.runSH(rec, req)

	body := rec.Body.String()
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status %d", rec.Code)
	}
	if !strings.Contains(body, "http://relay.example.com/download/agent/${os}/${arch}") {
		t.Fatalf("run.sh did not include agent download URL:\n%s", body)
	}
	if !strings.Contains(body, `exec "$tmp" --server "http://relay.example.com" --token "secret"`) {
		t.Fatalf("run.sh did not execute agent with expected server/token:\n%s", body)
	}
}

func TestRunSHDefaultsToRequestHost(t *testing.T) {
	app := NewApp(Config{})
	req := httptest.NewRequest(http.MethodGet, "http://lan.example.test:8080/run.sh", nil)
	rec := httptest.NewRecorder()

	app.runSH(rec, req)

	body := rec.Body.String()
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status %d", rec.Code)
	}
	if !strings.Contains(body, "http://lan.example.test:8080/download/agent/${os}/${arch}") {
		t.Fatalf("run.sh did not use request host for download URL:\n%s", body)
	}
	if !strings.Contains(body, `exec "$tmp" --server "http://lan.example.test:8080"`) {
		t.Fatalf("run.sh did not use request host for server URL:\n%s", body)
	}
}

func TestRunPS1UsesRunDownloadFlow(t *testing.T) {
	app := NewApp(Config{PublicHost: "relay.example.com"})
	req := httptest.NewRequest(http.MethodGet, "http://relay.example.com/run.ps1", nil)
	rec := httptest.NewRecorder()

	app.runPS1(rec, req)

	body := rec.Body.String()
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status %d", rec.Code)
	}
	if !strings.Contains(body, `$url = "http://relay.example.com/download/agent/windows/$arch"`) {
		t.Fatalf("run.ps1 did not include agent download URL:\n%s", body)
	}
	if !strings.Contains(body, `& $tmp --server "http://relay.example.com"`) {
		t.Fatalf("run.ps1 did not execute agent with expected server:\n%s", body)
	}
}

func TestRunPS1DefaultsToRequestHost(t *testing.T) {
	app := NewApp(Config{})
	req := httptest.NewRequest(http.MethodGet, "http://lan.example.test:8080/run.ps1", nil)
	rec := httptest.NewRecorder()

	app.runPS1(rec, req)

	body := rec.Body.String()
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status %d", rec.Code)
	}
	if !strings.Contains(body, `$url = "http://lan.example.test:8080/download/agent/windows/$arch"`) {
		t.Fatalf("run.ps1 did not use request host for download URL:\n%s", body)
	}
	if !strings.Contains(body, `& $tmp --server "http://lan.example.test:8080"`) {
		t.Fatalf("run.ps1 did not use request host for server URL:\n%s", body)
	}
}
