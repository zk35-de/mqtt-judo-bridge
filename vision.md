# judo2mqtt – Vision

## Why this exists

The Judo i-soft plus water softener has no open local API.
Its internal `DevCommManagerDaemon` speaks JSON over TCP port 8833 — reachable on the local network.
judo2mqtt connects to this daemon, receives push notifications, and publishes the values to MQTT
including Home Assistant Autodiscovery.

## Goal

Water softener data in Home Assistant — no cloud, no Judo account, no changes to the device.

## In scope

- TCP client for DevCommManagerDaemon port 8833 (JSON + 2-byte length-prefix framing)
- Login handshake: login → get devices → connect with want_notification
- MQTT publishing for relevant push groups: consumption, waterstop, settings, info
- Home Assistant MQTT Autodiscovery
- Docker container (linux/amd64, linux/arm64, linux/arm/v7)
- Reconnect logic on connection loss
- Web UI (prism-ui, statically embedded) for runtime configuration
- Valve control via MQTT command topic

## Out of scope

- Support for other Judo models (only i-soft plus tested)

## Architecture decisions

### Bidirectional DCM client

The DCM client is designed for bidirectional communication from the start —
even though early versions only read (receive push notifications).

**Why:** Whether the DCM supports write commands (e.g. open/close valve) was initially
unknown. Making the TCP writer mutex-protected from day one allows new commands to be
added as methods without refactoring the core structure.

**Concretely:** `sync.Mutex` on `send()` — write commands are new methods on top,
no structural change to the core.

### Valve control via MQTT

The valve (Wasserstop) can be controlled via MQTT command topic:

```
judo/switch/waterstop/set  →  ON / OFF
```

Home Assistant Autodiscovery registers this as a `switch` entity.

## Roadmap

- v0.1 – Connection + login handshake + all fields on MQTT
- v0.2 – HA Autodiscovery
- v0.3 – Reconnect logic, health endpoint, Docker Hub
- v0.4 – Valve control (waterstop open/close via MQTT)
- v0.5 – Web UI for runtime configuration
