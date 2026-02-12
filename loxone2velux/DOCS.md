# Loxone2Velux Gateway

## Über dieses Add-on

Dieses Add-on stellt eine Brücke zwischen Loxone Hausautomation und dem Velux
KLF-200 Gateway her. Es ermöglicht die Steuerung von Velux Fenstern, Rollläden
und Jalousien über Loxones HTTP Virtual Outputs oder das integrierte Web-Interface.

## Funktionen

- Steuerung von Velux Fenstern, Rollläden und Jalousien via REST API
- Web-Dashboard mit Echtzeit-Statusanzeige
- Loxone-kompatible HTTP-Endpunkte (einfache GET-Requests)
- Regen- und Windsensor-Status vom KLF-200
- Automatische Wiederverbindung bei Verbindungsverlust
- Persistente Konfiguration

## Konfiguration

### KLF-200 Einstellungen

- **klf200_host** (erforderlich): IP-Adresse oder Hostname des KLF-200
- **klf200_password** (erforderlich): WLAN-Passwort auf der Rückseite des KLF-200
- **klf200_port** (Standard: 51200): WebSocket-Port des KLF-200
- **reconnect_interval** (Standard: 30): Sekunden zwischen Reconnect-Versuchen
- **refresh_interval** (Standard: 300): Sekunden zwischen Status-Aktualisierungen

### Weitere Einstellungen

- **log_level** (Standard: info): Log-Level (debug, info, warn, error)
- **api_token** (optional): API-Token für Authentifizierung. Leer lassen um
  Authentifizierung zu deaktivieren.

## Web-Interface

Das Web-Interface ist über die Home Assistant Sidebar erreichbar (Ingress).
Das Dashboard zeigt alle verbundenen Velux Geräte mit Echtzeit-Positionsanzeige.
Geräte können geöffnet, geschlossen, gestoppt oder auf eine bestimmte Position
gefahren werden.

## Loxone Integration

Konfiguriere den Loxone Miniserver mit Virtual Outputs für folgende Endpunkte:

### Steuerung

| Aktion        | URL                                              |
| ------------- | ------------------------------------------------ |
| Öffnen        | `http://<HA_IP>:8080/loxone/node/{id}/open`      |
| Schliessen    | `http://<HA_IP>:8080/loxone/node/{id}/close`     |
| Stopp         | `http://<HA_IP>:8080/loxone/node/{id}/stop`      |
| Position      | `http://<HA_IP>:8080/loxone/node/{id}/set/{pct}` |

### Sensoren

| Daten         | URL                                              |
| ------------- | ------------------------------------------------ |
| Alle Sensoren | `http://<HA_IP>:8080/loxone/sensors`             |
| Nur Regen     | `http://<HA_IP>:8080/loxone/sensors/rain`        |
| Nur Wind      | `http://<HA_IP>:8080/loxone/sensors/wind`        |

Ersetze `{id}` mit der Velux Node-ID und `{pct}` mit 0-100.

Falls API-Token gesetzt, `?token=DEIN_TOKEN` an die URL anhängen.

## Netzwerk

Der KLF-200 muss vom Home Assistant Host auf Port 51200 (TCP/TLS) erreichbar
sein. Stelle sicher, dass deine Netzwerkkonfiguration dies erlaubt.

Port 8080 wird für den direkten Loxone-Zugriff auf dem Host exponiert.

## Support

Issues und Feature-Requests:
https://github.com/stefanbeyeler/loxone2velux/issues
