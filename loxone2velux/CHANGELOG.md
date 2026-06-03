# Changelog

## 1.2.0

- Fix: Positionsanzeige/UDP-Feedback wurde nach Run-Status-Meldungen fälschlich auf 0 % gesetzt
- Fix: Loxone-Mappings und UDP-Feedback überleben jetzt Add-on-Neustarts (separate `loxone.yaml`)
- Fix: `Stop`-Kommando wartet nun auf Bestätigung des KLF-200 und meldet Fehler
- Fix: Sensor-Refresh läuft nicht mehr in einen toten Sammel-Loop (schneller, korrekt)
- KLF-200-Operationen werden serialisiert, damit Antworten nicht vertauscht werden
- KLF-200-Einstellungsänderungen lösen automatisch einen Reconnect aus
- Sicherheit: KLF-200-Passwort wird nicht mehr im Debug-Log ausgegeben
- Interne Aufräumarbeiten (Data-Race im Reader, Config-Kopien, toter Code)

## 1.1.0

- UDP-Feedback an Loxone Miniserver (Echtzeit-Push von Position, State, Sensoren)
- Node-zu-Loxone Mapping-System mit CRUD-API (`/api/mappings`)
- Loxone-Konfiguration via API (`/api/loxone/config`)
- UDP-Test-Endpoint (`/api/loxone/config/udp/test`)
- Versionsnummer im Frontend-Header angezeigt

## 1.0.8

- Korrekter s6-overlay v3 longrun Service statt CMD/ENTRYPOINT
- HA Base Image wiederhergestellt (s6-overlay Kompatibilität)
- bashio durch /bin/sh + jq ersetzt für zuverlässiges Options-Parsing
- Liest HA-Optionen direkt aus /data/options.json

## 1.0.0

- Initial Home Assistant Add-on Release
- Web-Dashboard mit Echtzeit-Geräteüberwachung
- Loxone-kompatible REST API Endpunkte
- Regen- und Windsensor-Unterstützung
- Automatische KLF-200 Wiederverbindung
- Persistente Konfiguration
- Home Assistant Ingress Support
