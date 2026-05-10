package state

import (
	"testing"
	"time"
)

func TestUpdateAndSnapshot(t *testing.T) {
	s := New()
	s.SetMQTT(true)
	s.SetDCM(true)
	s.Update(map[string]string{
		"water_total":       "870331",
		"water_softened":    "520324",
		"salt_quantity":     "24600",
		"salt_range":        "217",
		"residual_hardness": "8",
		"valve":             "opened",
	})

	snap := s.Snapshot()
	if !snap.MQTTConnected {
		t.Error("MQTTConnected should be true")
	}
	if snap.WaterTotal != "870331" {
		t.Errorf("WaterTotal: got %q", snap.WaterTotal)
	}
	if snap.SaltRange != "217" {
		t.Errorf("SaltRange: got %q", snap.SaltRange)
	}
	if snap.Valve != "opened" {
		t.Errorf("Valve: got %q", snap.Valve)
	}
	if snap.LastPoll == nil {
		t.Error("LastPoll should be set after Update")
	}
	if snap.UptimeSeconds < 0 {
		t.Error("UptimeSeconds should be >= 0")
	}
}

func TestSnapshotBeforePoll(t *testing.T) {
	s := New()
	snap := s.Snapshot()
	if snap.LastPoll != nil {
		t.Error("LastPoll should be nil before first poll")
	}
}

func TestStartTime(t *testing.T) {
	before := time.Now()
	s := New()
	after := time.Now()
	snap := s.Snapshot()
	if snap.UptimeSeconds < 0 {
		t.Error("negative uptime")
	}
	_ = before
	_ = after
}
