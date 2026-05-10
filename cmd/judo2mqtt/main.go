package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	pahoMQTT "github.com/eclipse/paho.mqtt.golang"
	"git.zk35.de/secalpha/judo2mqtt/internal/config"
	"git.zk35.de/secalpha/judo2mqtt/internal/dcm"
	judoMQTT "git.zk35.de/secalpha/judo2mqtt/internal/mqtt"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config error", "err", err)
		os.Exit(1)
	}

	level := slog.LevelInfo
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})))
	slog.Info("judo2mqtt starting", "host", cfg.JudoHost, "serial", cfg.JudoSerial)

	// MQTT
	opts := pahoMQTT.NewClientOptions().
		AddBroker(cfg.MQTTBroker).
		SetClientID("judo2mqtt").
		SetAutoReconnect(true).
		SetOnConnectHandler(func(_ pahoMQTT.Client) {
			slog.Info("mqtt connected", "broker", cfg.MQTTBroker)
		})
	mqttClient := pahoMQTT.NewClient(opts)
	if tok := mqttClient.Connect(); tok.Wait() && tok.Error() != nil {
		slog.Error("mqtt connect failed", "err", tok.Error())
		os.Exit(1)
	}
	defer mqttClient.Disconnect(500)

	pub := judoMQTT.New(&pahoClientAdapter{mqttClient}, cfg.MQTTTopicPrefix, cfg.MQTTHADiscovery, cfg.MQTTHAPrefix)

	// DCM
	dcmClient := dcm.New(cfg.JudoHost, cfg.JudoPort, cfg.JudoUser, cfg.JudoSerial)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := dcmClient.Connect(ctx); err != nil {
		slog.Error("dcm connect failed", "err", err)
		os.Exit(1)
	}
	defer dcmClient.Close()

	pub.RegisterDiscovery(cfg.JudoSerial)

	ticker := time.NewTicker(time.Duration(cfg.PollIntervalSec) * time.Second)
	defer ticker.Stop()

	poll(ctx, dcmClient, pub)
	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down")
			return
		case <-ticker.C:
			poll(ctx, dcmClient, pub)
		}
	}
}

func poll(ctx context.Context, c *dcm.Client, pub *judoMQTT.Publisher) {
	data := map[string]string{}

	fetch := func(group, command, key string) {
		resp, err := c.Poll(ctx, group, command)
		if err != nil {
			slog.Warn("poll failed", "command", command, "err", err)
			return
		}
		if resp["status"] == "ok" {
			data[key] = strings.TrimSpace(fmt.Sprintf("%v", resp["data"]))
		}
	}

	// water total returns " raw softened" – split here
	resp, err := c.Poll(ctx, "consumption", "water total")
	if err == nil && resp["status"] == "ok" {
		parts := strings.Fields(fmt.Sprintf("%v", resp["data"]))
		if len(parts) == 2 {
			data["water_total"] = parts[0]
			data["water_softened"] = parts[1]
		}
	}

	fetch("consumption", "water average", "water_average")
	fetch("consumption", "salt quantity", "salt_quantity")
	fetch("consumption", "salt range", "salt_range")
	fetch("waterstop", "valve", "valve")
	fetch("settings", "residual hardness", "residual_hardness")

	pub.PublishAll(data)
	slog.Debug("polled", "fields", len(data))
}

// pahoClientAdapter adapts paho.mqtt.Client to our judoMQTT.Client interface.
type pahoClientAdapter struct {
	c pahoMQTT.Client
}

func (a *pahoClientAdapter) Publish(topic string, qos byte, retained bool, payload any) error {
	tok := a.c.Publish(topic, qos, retained, payload)
	tok.Wait()
	return tok.Error()
}
