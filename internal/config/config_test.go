package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingRequired(t *testing.T) {
	os.Unsetenv("JUDO_HOST")
	os.Unsetenv("JUDO_SERIAL")
	os.Unsetenv("CONFIG_FILE")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when required fields missing")
	}
}

func TestLoadDefaults(t *testing.T) {
	os.Setenv("JUDO_HOST", "10.35.5.133")
	os.Setenv("JUDO_SERIAL", "122907")
	defer os.Unsetenv("JUDO_HOST")
	defer os.Unsetenv("JUDO_SERIAL")

	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.JudoPort != 8833 {
		t.Errorf("JudoPort default: got %d", c.JudoPort)
	}
	if c.MQTTTopicPrefix != "judo" {
		t.Errorf("MQTTTopicPrefix default: got %s", c.MQTTTopicPrefix)
	}
	if !c.MQTTHADiscovery {
		t.Error("MQTTHADiscovery default should be true")
	}
	if c.PollIntervalSec != 60 {
		t.Errorf("PollIntervalSec default: got %d", c.PollIntervalSec)
	}
}

func TestLoadNoSecretLogging(t *testing.T) {
	os.Setenv("JUDO_HOST", "10.35.5.133")
	os.Setenv("JUDO_SERIAL", "122907")
	os.Setenv("JUDO_USER", "customer")
	defer os.Unsetenv("JUDO_HOST")
	defer os.Unsetenv("JUDO_SERIAL")
	defer os.Unsetenv("JUDO_USER")

	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.JudoUser != "customer" {
		t.Errorf("JudoUser: got %s", c.JudoUser)
	}
}

func TestLoadMQTTAuth(t *testing.T) {
	os.Setenv("JUDO_HOST", "x")
	os.Setenv("JUDO_SERIAL", "x")
	os.Setenv("MQTT_USER", "mqttuser")
	os.Setenv("MQTT_PASSWORD", "mqttha")
	defer os.Unsetenv("JUDO_HOST")
	defer os.Unsetenv("JUDO_SERIAL")
	defer os.Unsetenv("MQTT_USER")
	defer os.Unsetenv("MQTT_PASSWORD")

	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.MQTTUser != "mqttuser" {
		t.Errorf("MQTTUser: got %q", c.MQTTUser)
	}
	if c.MQTTPassword != "mqttha" {
		t.Errorf("MQTTPassword: got %q", c.MQTTPassword)
	}
}

func TestLoadHADiscoveryDisable(t *testing.T) {
	os.Setenv("JUDO_HOST", "x")
	os.Setenv("JUDO_SERIAL", "x")
	os.Setenv("MQTT_HA_DISCOVERY", "false")
	defer os.Unsetenv("JUDO_HOST")
	defer os.Unsetenv("JUDO_SERIAL")
	defer os.Unsetenv("MQTT_HA_DISCOVERY")

	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.MQTTHADiscovery {
		t.Error("MQTTHADiscovery should be false")
	}
}

func TestSaveAndLoadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "judo2mqtt.json")

	haDisc := true
	fc := FileConfig{
		JudoHost:        "192.168.1.50",
		JudoPort:        8833,
		JudoSerial:      "999999",
		MQTTBroker:      "tcp://mqtt.local:1883",
		MQTTUser:        "user",
		MQTTPassword:    "pass",
		MQTTTopicPrefix: "judo",
		MQTTHADiscovery: &haDisc,
		PollIntervalSec: 30,
	}

	if err := Save(path, fc); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.JudoHost != "192.168.1.50" {
		t.Errorf("JudoHost: got %q", loaded.JudoHost)
	}
	if loaded.MQTTPassword != "pass" {
		t.Errorf("MQTTPassword: got %q", loaded.MQTTPassword)
	}
	if loaded.MQTTHADiscovery == nil || !*loaded.MQTTHADiscovery {
		t.Error("MQTTHADiscovery: expected true")
	}
	if loaded.PollIntervalSec != 30 {
		t.Errorf("PollIntervalSec: got %d", loaded.PollIntervalSec)
	}
}

func TestFileConfigOverridesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "judo2mqtt.json")

	haDisc := false
	if err := Save(path, FileConfig{
		JudoHost:        "192.168.1.50",
		JudoSerial:      "999999",
		MQTTBroker:      "tcp://custom:1883",
		MQTTHADiscovery: &haDisc,
		PollIntervalSec: 120,
	}); err != nil {
		t.Fatal(err)
	}

	os.Setenv("CONFIG_FILE", path)
	os.Unsetenv("JUDO_HOST")
	os.Unsetenv("JUDO_SERIAL")
	defer os.Unsetenv("CONFIG_FILE")

	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.JudoHost != "192.168.1.50" {
		t.Errorf("JudoHost from file: got %q", c.JudoHost)
	}
	if c.MQTTBroker != "tcp://custom:1883" {
		t.Errorf("MQTTBroker from file: got %q", c.MQTTBroker)
	}
	if c.MQTTHADiscovery {
		t.Error("MQTTHADiscovery from file: expected false")
	}
	if c.PollIntervalSec != 120 {
		t.Errorf("PollIntervalSec from file: got %d", c.PollIntervalSec)
	}
}

func TestEnvVarOverridesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "judo2mqtt.json")

	if err := Save(path, FileConfig{
		JudoHost:   "192.168.1.50",
		JudoSerial: "999999",
	}); err != nil {
		t.Fatal(err)
	}

	os.Setenv("CONFIG_FILE", path)
	os.Setenv("JUDO_HOST", "10.0.0.1")
	defer os.Unsetenv("CONFIG_FILE")
	defer os.Unsetenv("JUDO_HOST")
	defer os.Unsetenv("JUDO_SERIAL")

	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.JudoHost != "10.0.0.1" {
		t.Errorf("env should override file: got %q", c.JudoHost)
	}
}
