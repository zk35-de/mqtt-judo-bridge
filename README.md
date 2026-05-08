# judo2mqtt

Judo i-soft plus → MQTT bridge.

Verbindet sich mit dem `DevCommManagerDaemon` (Port 8833) des Judo i-soft plus
und veröffentlicht Wasserverbrauchs-, Salz-, Regenerations- und Wasserstop-Daten
auf MQTT – inklusive Home Assistant Autodiscovery.

Keine Cloud. Keine Änderungen am Gerät. Kein Judo-Account.

## Voraussetzungen

- Judo i-soft plus im lokalen Netzwerk erreichbar
- MQTT Broker (z.B. Mosquitto)
- Docker

## Schnellstart

```bash
cp .env.example .env
# .env anpassen
docker compose up -d
```

## Konfiguration

| Variable | Beschreibung | Default |
|----------|-------------|---------|
| `JUDO_HOST` | IP/Hostname des Judo-Geräts | – |
| `JUDO_PORT` | DCM TCP-Port | `8833` |
| `JUDO_USER` | DCM Benutzername | `customer` |
| `MQTT_BROKER` | MQTT Broker URL | `tcp://localhost:1883` |
| `MQTT_TOPIC_PREFIX` | MQTT Topic Prefix | `judo` |
| `MQTT_HA_DISCOVERY` | Home Assistant Autodiscovery | `true` |
| `MQTT_HA_PREFIX` | HA Discovery Prefix | `homeassistant` |
| `LOG_LEVEL` | Log-Level (debug/info/warn/error) | `info` |

## MQTT Topics

```
judo/sensor/water_total
judo/sensor/water_current
judo/sensor/salt_quantity
judo/sensor/salt_range
judo/binary_sensor/regeneration_active
judo/binary_sensor/waterstop_open
judo/sensor/errors
```

## Build

```bash
go build ./...
docker build .
```
