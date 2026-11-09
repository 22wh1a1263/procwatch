package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func get(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestHealthz(t *testing.T) {
	if rec := get(t, "/healthz"); rec.Code != 200 || rec.Body.String() != "ok" {
		t.Fatalf("got %d %q", rec.Code, rec.Body.String())
	}
}

func TestSnapshotJSON(t *testing.T) {
	rec := get(t, "/api/snapshot")
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"load", "memory", "disk", "uptime_seconds"} {
		if _, ok := m[k]; !ok {
			t.Fatalf("missing key %q", k)
		}
	}
}

func TestTopLimitAndBadParam(t *testing.T) {
	var ps []map[string]any
	json.Unmarshal(get(t, "/api/top?n=2").Body.Bytes(), &ps)
	if len(ps) > 2 {
		t.Fatalf("n=2 returned %d", len(ps))
	}
	if rec := get(t, "/api/top?n=abc"); rec.Code != 200 {
		t.Fatalf("bad n should fall back to default, got %d", rec.Code)
	}
}

func TestIndexPage(t *testing.T) {
	rec := get(t, "/")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "procwatch") {
		t.Fatalf("index page not served: %d", rec.Code)
	}
}

func TestHandlerWithDiskUsesGivenPath(t *testing.T) {
	rec := httptest.NewRecorder()
	HandlerWithDisk("/tmp").ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/snapshot", nil))
	var m struct {
		Disk struct {
			Path string `json:"path"`
		} `json:"disk"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil || m.Disk.Path != "/tmp" {
		t.Fatalf("got %q, %v", m.Disk.Path, err)
	}
}

func TestEmptyListsEncodeAsArrays(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, []int{}, nil)
	if strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Fatalf("got %q", rec.Body.String())
	}
}
