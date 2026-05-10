package mqtt

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

// Client is the interface for an MQTT connection (allows mocking in tests).
type Client interface {
	Publish(topic string, qos byte, retained bool, payload any) error
}

// Publisher publishes Judo sensor values to an MQTT broker
// and registers Home Assistant Autodiscovery topics.
type Publisher struct {
	client      Client
	topicPrefix string
	haDiscovery bool
	haPrefix    string
}

func New(client Client, topicPrefix string, haDiscovery bool, haPrefix string) *Publisher {
	return &Publisher{
		client:      client,
		topicPrefix: topicPrefix,
		haDiscovery: haDiscovery,
		haPrefix:    haPrefix,
	}
}

type sensorDef struct {
	topic       string
	name        string
	unit        string
	deviceClass string
	stateClass  string
	icon        string
}

var sensors = []sensorDef{
	{"sensor/water_total", "Water Total", "L", "water", "total_increasing", ""},
	{"sensor/water_softened", "Water Softened", "L", "water", "total_increasing", ""},
	{"sensor/water_average", "Water Average Daily", "L", "", "measurement", "mdi:waves"},
	{"sensor/salt_quantity", "Salt Quantity", "g", "weight", "measurement", ""},
	{"sensor/salt_range", "Salt Range", "d", "duration", "measurement", ""},
	{"sensor/residual_hardness", "Residual Hardness", "°dH", "", "measurement", "mdi:water-check"},
}

var binarySensors = []struct {
	topic       string
	name        string
	deviceClass string
}{
	{"binary_sensor/waterstop_open", "Waterstop Valve", "opening"},
}

// PublishAll publishes all sensor values derived from polled data.
func (p *Publisher) PublishAll(data map[string]string) {
	p.publishSensor("sensor/water_total", data["water_total"])
	p.publishSensor("sensor/water_softened", data["water_softened"])
	p.publishSensor("sensor/water_average", data["water_average"])
	p.publishSensor("sensor/salt_quantity", data["salt_quantity"])
	p.publishSensor("sensor/salt_range", data["salt_range"])
	p.publishSensor("sensor/residual_hardness", data["residual_hardness"])
	p.publishBinary("binary_sensor/waterstop_open", data["valve"] == "opened")
}

func (p *Publisher) publishSensor(subtopic, value string) {
	if value == "" {
		return
	}
	topic := fmt.Sprintf("%s/%s", p.topicPrefix, subtopic)
	if err := p.client.Publish(topic, 0, true, value); err != nil {
		slog.Warn("mqtt publish failed", "topic", topic, "err", err)
	}
}

func (p *Publisher) publishBinary(subtopic string, state bool) {
	val := "OFF"
	if state {
		val = "ON"
	}
	topic := fmt.Sprintf("%s/%s", p.topicPrefix, subtopic)
	if err := p.client.Publish(topic, 0, true, val); err != nil {
		slog.Warn("mqtt publish failed", "topic", topic, "err", err)
	}
}

// RegisterDiscovery publishes Home Assistant MQTT Autodiscovery configs.
func (p *Publisher) RegisterDiscovery(deviceSerial string) {
	if !p.haDiscovery {
		return
	}
	device := map[string]any{
		"identifiers":  []string{"judo_" + deviceSerial},
		"name":         "Judo i-soft plus",
		"manufacturer": "JUDO Wasseraufbereitung",
		"model":        "i-soft plus",
		"serial_number": deviceSerial,
	}
	for _, s := range sensors {
		uid := "judo_" + deviceSerial + "_" + strings.ReplaceAll(s.topic, "/", "_")
		cfg := map[string]any{
			"unique_id":           uid,
			"name":                s.name,
			"state_topic":         fmt.Sprintf("%s/%s", p.topicPrefix, s.topic),
			"unit_of_measurement": s.unit,
			"device":              device,
		}
		if s.deviceClass != "" {
			cfg["device_class"] = s.deviceClass
		}
		if s.stateClass != "" {
			cfg["state_class"] = s.stateClass
		}
		if s.icon != "" {
			cfg["icon"] = s.icon
		}
		p.publishDiscovery("sensor", uid, cfg)
	}
	for _, b := range binarySensors {
		uid := "judo_" + deviceSerial + "_" + strings.ReplaceAll(b.topic, "/", "_")
		cfg := map[string]any{
			"unique_id":    uid,
			"name":         b.name,
			"state_topic":  fmt.Sprintf("%s/%s", p.topicPrefix, b.topic),
			"device_class": b.deviceClass,
			"device":       device,
		}
		p.publishDiscovery("binary_sensor", uid, cfg)
	}
}

func (p *Publisher) publishDiscovery(component, uid string, cfg map[string]any) {
	topic := fmt.Sprintf("%s/%s/%s/config", p.haPrefix, component, uid)
	payload, err := json.Marshal(cfg)
	if err != nil {
		slog.Warn("mqtt discovery marshal failed", "err", err)
		return
	}
	if err := p.client.Publish(topic, 0, true, string(payload)); err != nil {
		slog.Warn("mqtt discovery publish failed", "topic", topic, "err", err)
	}
}
