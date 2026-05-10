# judo2mqtt

Judo i-soft plus → MQTT bridge.

Verbindet sich mit dem `DevCommManagerDaemon` (Port 8833) des Judo i-soft plus
und veröffentlicht Wasserverbrauchs-, Salz- und Wasserstop-Daten auf MQTT –
inklusive Home Assistant Autodiscovery.

Keine Cloud. Keine Änderungen am Gerät. Kein Judo-Account.

## Schnellstart

```bash
mkdir config
docker compose up -d
```

Der Container startet auch ohne Konfiguration. Öffne die WebUI
(Standard: `http://localhost:8050`), trage Judo-IP und MQTT-Daten ein
und klicke **Speichern** – die Verbindung wird sofort aufgebaut.

## Konfiguration

### Option A: WebUI (empfohlen)

Konfiguration über `http://<host>:8050` → Karte **Konfiguration**.
Werte werden in `./config/judo2mqtt.json` gespeichert und überleben Container-Neustarts.

### Option B: Env-Vars / `.env`

Env-Vars haben Vorrang gegenüber der Config-Datei.

```bash
cp .env.example .env
# .env anpassen
docker compose up -d
```

| Variable | Beschreibung | Default |
|----------|-------------|---------|
| `JUDO_HOST` | IP/Hostname des Judo-Geräts | – |
| `JUDO_SERIAL` | Seriennummer (6-stellig) | – |
| `JUDO_PORT` | DCM TCP-Port | `8833` |
| `JUDO_USER` | DCM Benutzername | `customer` |
| `MQTT_BROKER` | MQTT Broker URL | `tcp://localhost:1883` |
| `MQTT_USER` | MQTT Benutzername | – |
| `MQTT_PASSWORD` | MQTT Passwort | – |
| `MQTT_TOPIC_PREFIX` | MQTT Topic Prefix | `judo` |
| `MQTT_HA_DISCOVERY` | Home Assistant Autodiscovery | `true` |
| `MQTT_HA_PREFIX` | HA Discovery Prefix | `homeassistant` |
| `POLL_INTERVAL` | Poll-Intervall in Sekunden | `60` |
| `WEB_ADDR` | WebUI Adresse | `:8080` |
| `CONFIG_FILE` | Pfad zur Config-Datei | `/config/judo2mqtt.json` |
| `LOG_LEVEL` | Log-Level (debug/info/warn/error) | `info` |

### Config-Datei

`/config/judo2mqtt.json` (im Container, via Volume-Mount):

```json
{
  "judo_host": "10.35.5.133",
  "judo_serial": "122907",
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
    # env_file: .env   # optional, Env-Vars überschreiben Config-Datei
```

## Build

```bash
go test ./...
go build ./...
podman build .
```

## Architektur

```
main
 ├── config.Load()          – Defaults → Config-Datei → Env-Vars
 ├── web.Server             – GET /api/status, GET/POST /api/config
 ├── dcm.Client             – TCP 8833, JSON + Length-Prefix
 ├── mqtt.Publisher         – Paho MQTT, HA Autodiscovery
 └── poll-loop              – Ticker + Hot-Reload Channel
```

**Hot-Reload:** POST /api/config speichert die Datei und sendet die neue
Config auf einen internen Channel. Main stoppt DCM + MQTT und baut die
Verbindungen mit den neuen Parametern neu auf – kein Container-Neustart.
