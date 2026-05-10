package main

import (
	"context"
	"log/slog"
	"os"
	"fmt"
	"os/signal"
	"strings"
	"syscall"
	"time"

	pahoMQTT "github.com/eclipse/paho.mqtt.golang"
	"git.zk35.de/secalpha/judo2mqtt/internal/config"
	"git.zk35.de/secalpha/judo2mqtt/internal/dcm"
	judoMQTT "git.zk35.de/secalpha/judo2mqtt/internal/mqtt"
	"git.zk35.de/secalpha/judo2mqtt/internal/state"
	"git.zk35.de/secalpha/judo2mqtt/internal/web"
)

const version = "v0.1.0"

type services struct {
	mqttClient pahoMQTT.Client
	dcmClient  *dcm.Client
	pub        *judoMQTT.Publisher
}

func (s *services) ready() bool {
	return s != nil && s.dcmClient != nil && s.pub != nil
}

func (s *services) stop(st *state.State) {
	if s.dcmClient != nil {
		s.dcmClient.Close()
		st.SetDCM(false)
	}
	if s.mqttClient != nil {
		s.mqttClient.Disconnect(500)
		st.SetMQTT(false)
	}
}

func startServices(ctx context.Context, cfg *config.Config, st *state.State) (*services, error) {
	opts := pahoMQTT.NewClientOptions().
		AddBroker(cfg.MQTTBroker).
		SetClientID("judo2mqtt").
		SetAutoReconnect(true).
		SetOnConnectHandler(func(_ pahoMQTT.Client) {
			slog.Info("mqtt connected", "broker", cfg.MQTTBroker)
			st.SetMQTT(true)
		}).
		SetConnectionLostHandler(func(_ pahoMQTT.Client, err error) {
			slog.Warn("mqtt connection lost", "err", err)
			st.SetMQTT(false)
		})
	if cfg.MQTTUser != "" {
		opts.SetUsername(cfg.MQTTUser)
		opts.SetPassword(cfg.MQTTPassword)
	}
	mqttClient := pahoMQTT.NewClient(opts)
	if tok := mqttClient.Connect(); tok.Wait() && tok.Error() != nil {
		return nil, tok.Error()
	}

	dcmClient := dcm.New(cfg.JudoHost, cfg.JudoPort, cfg.JudoUser, cfg.JudoSerial)
	if err := dcmClient.Connect(ctx); err != nil {
		mqttClient.Disconnect(500)
		return nil, err
	}
	st.SetDCM(true)

	pub := judoMQTT.New(&pahoClientAdapter{mqttClient}, cfg.MQTTTopicPrefix, cfg.MQTTHADiscovery, cfg.MQTTHAPrefix)
	pub.RegisterDiscovery(cfg.JudoSerial)

	return &services{mqttClient: mqttClient, dcmClient: dcmClient, pub: pub}, nil
}

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
	if cfg.IsComplete() {
		slog.Info("judo2mqtt starting", "host", cfg.JudoHost, "serial", cfg.JudoSerial, "version", version)
	} else {
		slog.Warn("judo2mqtt starting without config – open WebUI to configure", "addr", cfg.WebAddr, "version", version)
	}

	st := state.New()
	reloadCh := make(chan config.FileConfig, 1)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var svc *services
	if cfg.IsComplete() {
		svc, err = startServices(ctx, cfg, st)
		if err != nil {
			slog.Error("startup failed", "err", err)
			svc = &services{}
		}
	} else {
		svc = &services{}
	}

	webSrv := web.New(st, version, cfg, func(fc config.FileConfig) error {
		select {
		case reloadCh <- fc:
		default:
		}
		return nil
	})
	go func() {
		if err := webSrv.Start(ctx, cfg.WebAddr); err != nil {
			slog.Error("web server error", "err", err)
		}
	}()

	ticker := time.NewTicker(time.Duration(cfg.PollIntervalSec) * time.Second)
	defer ticker.Stop()

	if svc.ready() {
		poll(ctx, svc.dcmClient, svc.pub, st)
	}

	for {
		select {
		case <-ctx.Done():
			svc.stop(st)
			return

		case fc := <-reloadCh:
			slog.Info("config reload", "host", fc.JudoHost)
			svc.stop(st)
			config.Apply(cfg, fc)

			newSvc, err := startServices(ctx, cfg, st)
			if err != nil {
				slog.Error("reload failed, services stopped", "err", err)
				svc = &services{}
				continue
			}
			svc = newSvc
			ticker.Reset(time.Duration(cfg.PollIntervalSec) * time.Second)
			poll(ctx, svc.dcmClient, svc.pub, st)

		case <-ticker.C:
			if svc.ready() {
				poll(ctx, svc.dcmClient, svc.pub, st)
			}
		}
	}
}

func poll(ctx context.Context, c *dcm.Client, pub *judoMQTT.Publisher, st *state.State) {
	data := map[string]string{}

	fetch := func(group, command, key string) {
		resp, err := c.Poll(ctx, group, command)
		if err != nil {
			slog.Warn("poll failed", "command", command, "err", err)
			return
		}
		if resp["status"] == "ok" {
			val := resp["data"]
			if s, ok := val.(string); ok {
				data[key] = strings.TrimSpace(s)
			}
		}
	}

	resp, err := c.Poll(ctx, "consumption", "water total")
	if err == nil && resp["status"] == "ok" {
		parts := strings.Fields(strings.TrimSpace(fmt.Sprintf("%v", resp["data"])))
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
	st.Update(data)
	slog.Debug("polled", "fields", len(data))
}

type pahoClientAdapter struct {
	c pahoMQTT.Client
}

func (a *pahoClientAdapter) Publish(topic string, qos byte, retained bool, payload any) error {
	tok := a.c.Publish(topic, qos, retained, payload)
	tok.Wait()
	return tok.Error()
}
