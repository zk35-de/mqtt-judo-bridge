package state

import (
	"sync"
	"time"
)

// State holds the latest polled sensor values and connection flags.
// It is safe for concurrent use.
type State struct {
	mu sync.RWMutex

	MQTTConnected bool
	DCMConnected  bool
	LastPoll      time.Time
	StartTime     time.Time

	// Sensor values as last received from the device.
	WaterTotal       string
	WaterSoftened    string
	WaterAverage     string
	SaltQuantity     string
	SaltRange        string
	ResidualHardness string
	Valve            string
}

func New() *State {
	return &State{StartTime: time.Now()}
}

func (s *State) SetMQTT(connected bool) {
	s.mu.Lock()
	s.MQTTConnected = connected
	s.mu.Unlock()
}

func (s *State) SetDCM(connected bool) {
	s.mu.Lock()
	s.DCMConnected = connected
	s.mu.Unlock()
}

func (s *State) Update(data map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastPoll = time.Now()
	if v, ok := data["water_total"]; ok {
		s.WaterTotal = v
	}
	if v, ok := data["water_softened"]; ok {
		s.WaterSoftened = v
	}
	if v, ok := data["water_average"]; ok {
		s.WaterAverage = v
	}
	if v, ok := data["salt_quantity"]; ok {
		s.SaltQuantity = v
	}
	if v, ok := data["salt_range"]; ok {
		s.SaltRange = v
	}
	if v, ok := data["residual_hardness"]; ok {
		s.ResidualHardness = v
	}
	if v, ok := data["valve"]; ok {
		s.Valve = v
	}
}

type Snapshot struct {
	MQTTConnected    bool    `json:"mqtt_connected"`
	DCMConnected     bool    `json:"dcm_connected"`
	LastPoll         *string `json:"last_poll"`
	UptimeSeconds    int64   `json:"uptime_seconds"`
	WaterTotal       string  `json:"water_total_l"`
	WaterSoftened    string  `json:"water_softened_l"`
	WaterAverage     string  `json:"water_average_l_day"`
	SaltQuantity     string  `json:"salt_quantity_g"`
	SaltRange        string  `json:"salt_range_days"`
	ResidualHardness string  `json:"residual_hardness_ddh"`
	Valve            string  `json:"valve"`
}

func (s *State) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap := Snapshot{
		MQTTConnected:    s.MQTTConnected,
		DCMConnected:     s.DCMConnected,
		UptimeSeconds:    int64(time.Since(s.StartTime).Seconds()),
		WaterTotal:       s.WaterTotal,
		WaterSoftened:    s.WaterSoftened,
		WaterAverage:     s.WaterAverage,
		SaltQuantity:     s.SaltQuantity,
		SaltRange:        s.SaltRange,
		ResidualHardness: s.ResidualHardness,
		Valve:            s.Valve,
	}
	if !s.LastPoll.IsZero() {
		t := s.LastPoll.UTC().Format(time.RFC3339)
		snap.LastPoll = &t
	}
	return snap
}
