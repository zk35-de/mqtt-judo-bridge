package config

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
)

const DefaultConfigFile = "/config/judo2mqtt.json"

type Config struct {
	JudoHost   string
	JudoPort   int
	JudoUser   string
	JudoSerial string

	MQTTBroker      string
	MQTTUser        string
	MQTTPassword    string
	MQTTTopicPrefix string
	MQTTHADiscovery bool
	MQTTHAPrefix    string

	PollIntervalSec int
	WebAddr         string
	LogLevel        string

	ConfigFile string
}

// FileConfig is the JSON-persisted subset (env vars override on load).
type FileConfig struct {
	JudoHost        string `json:"judo_host,omitempty"`
	JudoPort        int    `json:"judo_port,omitempty"`
	JudoSerial      string `json:"judo_serial,omitempty"`
	MQTTBroker      string `json:"mqtt_broker,omitempty"`
	MQTTUser        string `json:"mqtt_user,omitempty"`
	MQTTPassword    string `json:"mqtt_password,omitempty"`
	MQTTTopicPrefix string `json:"mqtt_topic_prefix,omitempty"`
	MQTTHADiscovery *bool  `json:"mqtt_ha_discovery,omitempty"`
	MQTTHAPrefix    string `json:"mqtt_ha_prefix,omitempty"`
	PollIntervalSec int    `json:"poll_interval_sec,omitempty"`
}

func defaults() *Config {
	return &Config{
		JudoPort:        8833,
		JudoUser:        "customer",
		MQTTBroker:      "tcp://localhost:1883",
		MQTTTopicPrefix: "judo",
		MQTTHADiscovery: true,
		MQTTHAPrefix:    "homeassistant",
		PollIntervalSec: 60,
		WebAddr:         ":8080",
		LogLevel:        "info",
		ConfigFile:      DefaultConfigFile,
	}
}

func Load() (*Config, error) {
	c := defaults()

	if v := os.Getenv("CONFIG_FILE"); v != "" {
		c.ConfigFile = v
	}

	if fc, err := LoadFile(c.ConfigFile); err == nil {
		applyFileConfig(c, fc)
	}

	if v := os.Getenv("JUDO_HOST"); v != "" {
		c.JudoHost = v
	}
	if v := os.Getenv("JUDO_SERIAL"); v != "" {
		c.JudoSerial = v
	}
	if v := os.Getenv("JUDO_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			c.JudoPort = p
		}
	}
	if v := os.Getenv("JUDO_USER"); v != "" {
		c.JudoUser = v
	}
	if v := os.Getenv("MQTT_BROKER"); v != "" {
		c.MQTTBroker = v
	}
	c.MQTTUser = firstNonEmpty(os.Getenv("MQTT_USER"), c.MQTTUser)
	c.MQTTPassword = firstNonEmpty(os.Getenv("MQTT_PASSWORD"), c.MQTTPassword)
	if v := os.Getenv("MQTT_TOPIC_PREFIX"); v != "" {
		c.MQTTTopicPrefix = v
	}
	if v := os.Getenv("MQTT_HA_DISCOVERY"); v == "false" || v == "0" {
		c.MQTTHADiscovery = false
	}
	if v := os.Getenv("MQTT_HA_PREFIX"); v != "" {
		c.MQTTHAPrefix = v
	}
	if v := os.Getenv("POLL_INTERVAL"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			c.PollIntervalSec = i
		}
	}
	if v := os.Getenv("WEB_ADDR"); v != "" {
		c.WebAddr = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		c.LogLevel = v
	}

	var errs []string
	if c.JudoHost == "" {
		errs = append(errs, "JUDO_HOST is required")
	}
	if c.JudoSerial == "" {
		errs = append(errs, "JUDO_SERIAL is required")
	}
	if len(errs) > 0 {
		msg := ""
		for _, e := range errs {
			msg += e + "\n"
		}
		return nil, errors.New(msg)
	}
	return c, nil
}

func LoadFile(path string) (*FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var fc FileConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, err
	}
	return &fc, nil
}

// Save writes fc to path atomically.
func Save(path string, fc FileConfig) error {
	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// ToFileConfig extracts the persisted fields from a running Config.
func ToFileConfig(c *Config) FileConfig {
	haDisc := c.MQTTHADiscovery
	return FileConfig{
		JudoHost:        c.JudoHost,
		JudoPort:        c.JudoPort,
		JudoSerial:      c.JudoSerial,
		MQTTBroker:      c.MQTTBroker,
		MQTTUser:        c.MQTTUser,
		MQTTPassword:    c.MQTTPassword,
		MQTTTopicPrefix: c.MQTTTopicPrefix,
		MQTTHADiscovery: &haDisc,
		MQTTHAPrefix:    c.MQTTHAPrefix,
		PollIntervalSec: c.PollIntervalSec,
	}
}

func applyFileConfig(c *Config, fc *FileConfig) {
	if fc.JudoHost != "" {
		c.JudoHost = fc.JudoHost
	}
	if fc.JudoPort > 0 {
		c.JudoPort = fc.JudoPort
	}
	if fc.JudoSerial != "" {
		c.JudoSerial = fc.JudoSerial
	}
	if fc.MQTTBroker != "" {
		c.MQTTBroker = fc.MQTTBroker
	}
	if fc.MQTTUser != "" {
		c.MQTTUser = fc.MQTTUser
	}
	if fc.MQTTPassword != "" {
		c.MQTTPassword = fc.MQTTPassword
	}
	if fc.MQTTTopicPrefix != "" {
		c.MQTTTopicPrefix = fc.MQTTTopicPrefix
	}
	if fc.MQTTHADiscovery != nil {
		c.MQTTHADiscovery = *fc.MQTTHADiscovery
	}
	if fc.MQTTHAPrefix != "" {
		c.MQTTHAPrefix = fc.MQTTHAPrefix
	}
	if fc.PollIntervalSec > 0 {
		c.PollIntervalSec = fc.PollIntervalSec
	}
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
