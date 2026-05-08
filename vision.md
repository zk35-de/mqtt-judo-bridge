# judo2mqtt – Vision

## Wozu

Judo i-soft plus Wasserenthärter hat keine offene lokale API.
Der interne `DevCommManagerDaemon` spricht JSON über TCP Port 8833 (netzwerkweit erreichbar).
judo2mqtt verbindet sich mit diesem Daemon, empfängt Push-Notifications und veröffentlicht
die Werte auf MQTT – inklusive Home Assistant Autodiscovery.

## Ziel

Wasserenthärter-Daten in Home Assistant ohne Cloud, ohne Judo-Account, ohne Änderungen am Gerät.

## In-Scope

- TCP-Client für DevCommManagerDaemon Port 8833 (JSON + 2-Byte Length Prefix Framing)
- Login-Handshake: login → get devices → connect mit want notification
- MQTT Publishing für relevante Push-Gruppen: consumption, waterstop, settings, info
- Home Assistant MQTT Autodiscovery
- Web-UI (prism-ui, statisch eingebettet) zur Konfiguration: MQTT-Broker, Judo-Host/Port, Topic-Prefix, HA-Discovery
- Docker Container (linux/amd64, linux/arm64, linux/arm/v7)
- Reconnect-Logik bei Verbindungsverlust

## Out-of-Scope

- Schreib-Zugriff auf das Gerät (Einstellungen ändern, Regeneration auslösen)
- Web-UI
- Unterstützung anderer Judo-Modelle (nur i-soft plus getestet)

## Roadmap

- v0.1 – Verbindung + Login-Handshake + alle Felder auf MQTT
- v0.2 – HA Autodiscovery
- v0.3 – Reconnect-Logik, Health-Endpoint, Docker Hub
