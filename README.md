# judo2mqtt

Judo i-soft plus → MQTT bridge.

Connects to the `DevCommManagerDaemon` (port 8833) of the Judo i-soft plus
and publishes water consumption, salt, and valve data to MQTT —
including Home Assistant Autodiscovery.

## Local-first

No cloud. No Judo account. No changes to the device.
The bridge talks to an internal TCP daemon that is already running on your device.

→ [vision.md](vision.md) for the full rationale.

## Quickstart

```bash
mkdir config
docker compose up -d
```

The container starts without any configuration. Open the web UI
(default: `http://localhost:8050`), enter your Judo IP and MQTT details,
and click **Save** — the connection is established immediately.

## Configuration

### Option A: Web UI (recommended)

Configure via `http://<host>:8050` → **Configuration** card.
Values are stored in `./config/judo2mqtt.json` and survive container restarts.

### Option B: Environment variables / `.env`

Env vars take precedence over the config file.

```bash
cp .env.example .env
# edit .env
docker compose up -d
```

| Variable | Description | Default |
|----------|-------------|---------|
| `JUDO_HOST` | IP/hostname of the Judo device | – |
| `JUDO_SERIAL` | Serial number (6 digits) | – |
| `JUDO_PORT` | DCM TCP port | `8833` |
| `JUDO_USER` | DCM username | `customer` |
| `MQTT_BROKER` | MQTT broker URL | `tcp://localhost:1883` |
| `MQTT_USER` | MQTT username | – |
| `MQTT_PASSWORD` | MQTT password | – |
| `MQTT_TOPIC_PREFIX` | MQTT topic prefix | `judo` |
| `MQTT_HA_DISCOVERY` | Home Assistant Autodiscovery | `true` |
| `MQTT_HA_PREFIX` | HA discovery prefix | `homeassistant` |
| `POLL_INTERVAL` | Poll interval in seconds | `60` |
| `WEB_ADDR` | Web UI address | `:8080` |
| `CONFIG_FILE` | Path to config file | `/config/judo2mqtt.json` |
| `LOG_LEVEL` | Log level (debug/info/warn/error) | `info` |

### Config file

`/config/judo2mqtt.json` (in the container, via volume mount):

```json
{
  "judo_host": "192.168.1.x",
  "judo_serial": "XXXXXX",
  "mqtt_broker": "tcp://ha.local:1883",
  "mqtt_user": "mqttuser",
  "mqtt_password": "secret",
  "mqtt_topic_prefix": "judo",
  "mqtt_ha_discovery": true,
  "poll_interval_sec": 60
}
```

## MQTT Topics

```
judo/sensor/water_total_l
judo/sensor/water_softened_l
judo/sensor/water_average_l_day
judo/sensor/salt_quantity_g
judo/sensor/salt_range_days
judo/sensor/residual_hardness_ddh
judo/sensor/valve
judo/switch/waterstop/set   ← command topic (ON / OFF)
```

## docker-compose.yml

```yaml
services:
  judo2mqtt:
    image: git.zk35.de/secalpha/judo2mqtt:latest
    restart: unless-stopped
    ports:
      - "8050:8080"
    volumes:
      - ./config:/config
    # env_file: .env   # optional, env vars override config file
```

## Build

```bash
go test ./...
go build ./...
podman build .
```

## Architecture

```
main
 ├── config.Load()          – Defaults → config file → env vars
 ├── web.Server             – GET /api/status, GET/POST /api/config
 ├── dcm.Client             – TCP 8833, JSON + length-prefix framing
 ├── mqtt.Publisher         – Paho MQTT, HA Autodiscovery
 └── poll-loop              – Ticker + hot-reload channel
```

**Hot-reload:** POST /api/config saves the file and sends the new config on an
internal channel. Main stops DCM + MQTT and rebuilds the connections with the
new parameters — no container restart needed.

## License

MIT
