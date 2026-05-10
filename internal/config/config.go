package config

import (
	"errors"
	"os"
	"strconv"
)

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
}

func Load() (*Config, error) {
	c := &Config{
		JudoPort:        8833,
		JudoUser:        "customer",
		MQTTBroker:      "tcp://localhost:1883",
		MQTTTopicPrefix: "judo",
		MQTTHADiscovery: true,
		MQTTHAPrefix:    "homeassistant",
		PollIntervalSec: 60,
		WebAddr:         ":8080",
		LogLevel:        "info",
	}

	c.JudoHost = os.Getenv("JUDO_HOST")
	c.JudoSerial = os.Getenv("JUDO_SERIAL")

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
	c.MQTTUser = os.Getenv("MQTT_USER")
	c.MQTTPassword = os.Getenv("MQTT_PASSWORD")
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
	return c, nil
}
