package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.zk35.de/secalpha/judo2mqtt/internal/state"
)

func newTestServer() *Server {
	st := state.New()
	st.SetMQTT(true)
	st.SetDCM(true)
	st.Update(map[string]string{
		"salt_quantity": "24600",
		"valve":         "opened",
	})
	return New(st, "test")
}

func TestHandleHealth(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	s.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("health: got %d", w.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Errorf("status: got %q", body["status"])
	}
	if body["version"] != "test" {
		t.Errorf("version: got %q", body["version"])
	}
}

func TestHandleStatus(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	s.handleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d", w.Code)
	}
	var snap map[string]any
	if err := json.NewDecoder(w.Body).Decode(&snap); err != nil {
		t.Fatal(err)
	}
	if snap["mqtt_connected"] != true {
		t.Errorf("mqtt_connected: got %v", snap["mqtt_connected"])
	}
	if snap["salt_quantity_g"] != "24600" {
		t.Errorf("salt_quantity_g: got %v", snap["salt_quantity_g"])
	}
}

func TestHandleUI(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	s.handleUI(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ui: got %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("content-type: got %q", ct)
	}
}
