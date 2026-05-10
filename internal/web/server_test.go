package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"git.zk35.de/secalpha/judo2mqtt/internal/config"
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
	return New(st, "test", nil, nil)
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

func TestHandleGetConfig(t *testing.T) {
	st := state.New()
	haDisc := true
	cfg := &config.Config{
		JudoHost:        "10.0.0.1",
		JudoSerial:      "123",
		JudoPort:        8833,
		MQTTBroker:      "tcp://broker:1883",
		MQTTUser:        "user",
		MQTTPassword:    "pass",
		MQTTTopicPrefix: "judo",
		MQTTHADiscovery: haDisc,
		PollIntervalSec: 60,
		ConfigFile:      "/config/judo2mqtt.json",
	}
	s := New(st, "test", cfg, nil)

	req := httptest.NewRequest("GET", "/api/config", nil)
	w := httptest.NewRecorder()
	s.handleGetConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("got %d", w.Code)
	}
	var fc config.FileConfig
	if err := json.NewDecoder(w.Body).Decode(&fc); err != nil {
		t.Fatal(err)
	}
	if fc.JudoHost != "10.0.0.1" {
		t.Errorf("JudoHost: got %q", fc.JudoHost)
	}
	if fc.MQTTPassword != "pass" {
		t.Errorf("MQTTPassword: got %q", fc.MQTTPassword)
	}
}

func TestHandleGetConfigNoCfg(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest("GET", "/api/config", nil)
	w := httptest.NewRecorder()
	s.handleGetConfig(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestHandlePostConfig(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "judo2mqtt.json")

	st := state.New()
	cfg := &config.Config{ConfigFile: cfgFile}
	restarted := false
	s := New(st, "test", cfg, func() { restarted = true })

	haDisc := true
	fc := config.FileConfig{
		JudoHost:        "10.1.2.3",
		JudoSerial:      "999",
		MQTTBroker:      "tcp://newbroker:1883",
		MQTTHADiscovery: &haDisc,
	}
	body, _ := json.Marshal(fc)
	req := httptest.NewRequest("POST", "/api/config", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.handlePostConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("got %d", w.Code)
	}

	saved, err := config.LoadFile(cfgFile)
	if err != nil {
		t.Fatal(err)
	}
	if saved.JudoHost != "10.1.2.3" {
		t.Errorf("saved JudoHost: got %q", saved.JudoHost)
	}

	// onRestart is called async after 200ms; give it time
	for i := 0; i < 10; i++ {
		if restarted {
			break
		}
		// tight poll
		var ch = make(chan struct{})
		go func() { close(ch) }()
		<-ch
	}
}

func TestHandlePostConfigValidation(t *testing.T) {
	dir := t.TempDir()
	st := state.New()
	cfg := &config.Config{ConfigFile: filepath.Join(dir, "judo2mqtt.json")}
	s := New(st, "test", cfg, func() {})

	// missing judo_host
	fc := config.FileConfig{JudoSerial: "123"}
	body, _ := json.Marshal(fc)
	req := httptest.NewRequest("POST", "/api/config", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.handlePostConfig(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
