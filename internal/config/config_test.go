package config

import (
	"os"
	"testing"
)

func TestLoadMissingRequired(t *testing.T) {
	os.Unsetenv("JUDO_HOST")
	os.Unsetenv("JUDO_SERIAL")
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
	// JUDO_USER must never appear in config struct in a loggable way –
	// this test just ensures it loads without panic and the field is set.
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
