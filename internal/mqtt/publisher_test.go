package mqtt

import (
	"fmt"
	"strings"
	"testing"
)

type mockClient struct {
	published map[string]string
}

func (m *mockClient) Publish(topic string, _ byte, _ bool, payload any) error {
	if m.published == nil {
		m.published = map[string]string{}
	}
	m.published[topic] = fmt.Sprintf("%v", payload)
	return nil
}

func newMock() (*Publisher, *mockClient) {
	c := &mockClient{}
	return New(c, "judo", true, "homeassistant"), c
}

func TestPublishAll(t *testing.T) {
	p, c := newMock()
	p.PublishAll(map[string]string{
		"water_total":       "870331",
		"water_softened":    "520324",
		"water_average":     "283",
		"salt_quantity":     "24600",
		"salt_range":        "217",
		"residual_hardness": "8",
		"valve":             "opened",
	})

	checks := map[string]string{
		"judo/sensor/water_total":           "870331",
		"judo/sensor/salt_quantity":         "24600",
		"judo/sensor/salt_range":            "217",
		"judo/binary_sensor/waterstop_open": "ON",
	}
	for topic, want := range checks {
		if got := c.published[topic]; got != want {
			t.Errorf("topic %s: got %q want %q", topic, got, want)
		}
	}
}

func TestPublishAllValveClosed(t *testing.T) {
	p, c := newMock()
	p.PublishAll(map[string]string{"valve": "closed"})
	if got := c.published["judo/binary_sensor/waterstop_open"]; got != "OFF" {
		t.Errorf("valve closed: got %q", got)
	}
}

func TestRegisterDiscovery(t *testing.T) {
	p, c := newMock()
	p.RegisterDiscovery("122907")

	found := false
	for topic, payload := range c.published {
		if strings.HasPrefix(topic, "homeassistant/sensor/") && strings.Contains(payload, "122907") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected HA discovery topic with serial 122907")
	}
}

func TestRegisterDiscoveryDisabled(t *testing.T) {
	c := &mockClient{}
	p := New(c, "judo", false, "homeassistant")
	p.RegisterDiscovery("122907")
	if len(c.published) != 0 {
		t.Error("expected no publish when HA discovery disabled")
	}
}
