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
- Docker Container (linux/amd64, linux/arm64, linux/arm/v7)
- Reconnect-Logik bei Verbindungsverlust

## Out-of-Scope (vorerst)

- Web-UI (prism-ui, statisch eingebettet) zur Konfiguration
- Unterstützung anderer Judo-Modelle (nur i-soft plus getestet)

## Architektur-Entscheidungen

### Bidirektionaler DCM-Client (ab v0.1)

Der DCM-Client wird von Anfang an für bidirektionale Kommunikation ausgelegt –
auch wenn v0.1 nur lesend arbeitet (Push-Notifications empfangen).

**Warum:** Ob der DCM Schreib-Befehle unterstützt (z.B. Wasserstop öffnen/schließen)
ist noch nicht analysiert. Falls ja, soll das ohne strukturellen Umbau nachrüstbar sein.

**Konkret:** Der TCP-Writer wird mutex-gesichert gekapselt (`sync.Mutex` auf `send()`).
Schreib-Befehle können dann als neue Methoden draufgesetzt werden – kein Refactor der Kernstruktur.

**Protokoll-Analyse:** Beim Aufbau von #2 (DCM-Client) via Wireshark prüfen,
welche Nachrichten die Judo-App beim Wasserstop-Toggle sendet.

### Möglicher v0.x: Wasserstop-Steuerung via MQTT

Falls Protokoll-Analyse positiv:
- MQTT-Command-Topic: `judo/switch/waterstop/set` (Payload: `ON`/`OFF`)
- HA-Autodiscovery als `switch`-Entity
- Kein eigenes Issue bis Protokoll-Analyse abgeschlossen

## Roadmap

- v0.1 – Verbindung + Login-Handshake + alle Felder auf MQTT
- v0.2 – HA Autodiscovery
- v0.3 – Reconnect-Logik, Health-Endpoint, Docker Hub
