# Changelog

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
