# Code Review: Loxone2velux

**Datum:** 2026-04-18
**Branch:** claude/review-project-improvements-soBS7
**Umfang:** Gesamtes Projekt (Go-Backend, Konfiguration, Docker, Web-Frontend)

## Projektübersicht

Go-basierter Gateway (Go 1.21, ~3.718 Zeilen) zwischen Loxone Home-Automation und Velux KLF-200. Frontend in Vue/Vite/Tailwind.

**Kritischer Gesamtbefund:** Es sind keine Unit-Tests (`*_test.go`) im Projekt vorhanden.

### Struktur

```
/home/user/Loxone2velux/
├── cmd/gateway/main.go           # Einstiegspunkt
├── internal/
│   ├── api/                      # REST-API Server
│   │   ├── server.go
│   │   ├── handlers.go
│   │   └── middleware.go
│   ├── klf200/                   # KLF-200 Client
│   │   ├── client.go
│   │   ├── types.go
│   │   ├── commands.go
│   │   ├── nodes.go
│   │   └── ca.go
│   ├── config/config.go
│   ├── gateway/service.go
│   └── loxone/
│       ├── udp.go
│       └── mapping.go
├── web/                          # Vue/Vite/Tailwind Frontend
├── Dockerfile
├── docker-compose.yml
├── config.example.yaml
└── go.mod (Go 1.21)
```

---

## Kritische Sicherheitsprobleme

### 1. Passwort wird im Log im Klartext ausgegeben

- **Datei:** `internal/klf200/client.go:185`
- **Code:**
  ```go
  c.logger.Debug().
      Hex("frame", frame).
      Int("len", len(frame)).
      Str("password", c.password).
      Msg("Sending password frame")
  ```
- **Problem:** Das KLF-200-Passwort wird im Debug-Log im Klartext ausgegeben.
- **Empfehlung:** `Str("password", c.password)` entfernen oder auf `"***"` setzen.

### 2. Passwort wird über `GET /api/config` zurückgegeben

- **Datei:** `internal/api/handlers.go:421`
- **Code:**
  ```go
  KLF200: ConfigKLF200{
      ...
      Password: cfg.KLF200.Password,
      ...
  }
  ```
- **Problem:** Das Passwort ist für jeden API-Token-Inhaber auslesbar.
- **Empfehlung:** In Response maskieren (`"***"`) oder nur bei expliziter Anfrage zurückgeben.

### 3. Keine Begrenzung der Request-Body-Größe

- **Datei:** `internal/api/handlers.go:126, 451, 554, 582, 655`
- **Code:**
  ```go
  json.NewDecoder(r.Body).Decode(&req)
  ```
- **Problem:** Ohne `http.MaxBytesReader` kann ein Angreifer beliebig große Payloads senden (DoS durch Speichererschöpfung).
- **Empfehlung:**
  ```go
  r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MiB
  ```

---

## Hohe Priorität

### 4. `InsecureSkipVerify: true` ohne Custom-Validator

- **Datei:** `internal/klf200/client.go:126`
- **Code:**
  ```go
  tlsConfig := &tls.Config{
      RootCAs:            certPool,
      InsecureSkipVerify: true,
      ...
  }
  ```
- **Problem:** Hostname-Verifikation ist komplett deaktiviert. Nur die CA wird vorab geladen, aber ohne `VerifyPeerCertificate`-Callback wird die Chain nicht tatsächlich gegen die Velux-CA validiert → MITM möglich.
- **Empfehlung:** `VerifyPeerCertificate` implementieren, das die Kette gegen `certPool` prüft (Hostname-Check kann übersprungen werden, da Velux kein CN/SAN setzt).

### 5. `CORSMiddleware` definiert, aber nicht aktiviert

- **Datei:** `internal/api/middleware.go:40-53` vs. `internal/api/server.go`
- **Problem:** Middleware existiert, wird im Router aber nicht registriert. Entweder toter Code oder vergessen.
- **Empfehlung:** Entweder mit Whitelist aktivieren oder komplett entfernen.

### 6. UDP-Send-Fehler werden still verschluckt

- **Datei:** `internal/loxone/udp.go:70-86`
- **Problem:** `Send()` loggt Fehler, gibt aber keinen Fehler an den Aufrufer zurück. Ausfälle im Feedback-Pfad bleiben unbemerkt.
- **Empfehlung:** Fehler zurückgeben und im Aufrufer behandeln oder Metriken erfassen.

### 7. Fehlende `Content-Type`-Header bei Loxone-Endpoints

- **Dateien:** `internal/api/handlers.go:340-366` (LoxoneRainStatus, LoxoneWindStatus, LoxoneSensorStatus)
- **Code:** `w.Write([]byte("1"))` ohne Header.
- **Empfehlung:**
  ```go
  w.Header().Set("Content-Type", "text/plain; charset=utf-8")
  ```

---

## Mittlere Priorität

### 8. Context-Propagation in Loops

- **Datei:** `internal/gateway/service.go:135, 159`
- **Problem:** `refreshLoop()` und `reconnectLoop()` erstellen eigene `context.Background()` statt den Service-Context zu übernehmen. Graceful Shutdown wird dadurch unzuverlässig.
- **Empfehlung:** Parent-Context konsequent propagieren.

### 9. Keine UUID-Validierung für Mappings

- **Datei:** `internal/api/handlers.go:579, 613`
- **Problem:** Jeder String wird als UUID akzeptiert.
- **Empfehlung:** Mit `regexp` oder `github.com/google/uuid` validieren.

### 10. `GetConfig()` gibt Pointer statt Kopie zurück

- **Datei:** `cmd/gateway/main.go:42-90`
- **Problem:** Race-Conditions möglich, wenn Aufrufer die Config außerhalb des Locks modifiziert.
- **Empfehlung:** Deep-Copy zurückgeben.

### 11. Server-Start-Fehler nicht an `main()` propagiert

- **Datei:** `cmd/gateway/main.go:159-163`
- **Problem:** Der API-Server läuft in einer Goroutine; Fehler beim Binden enden nur in `log.Fatal`. Der Hauptprozess registriert den Fehler nicht strukturiert.
- **Empfehlung:** Error-Channel zwischen Goroutine und `main()` einführen.

### 12. Kein Rate-Limiting

- **Dateien:** `internal/api/server.go`, `internal/api/middleware.go`
- **Problem:** API ist ohne Rate-Limit exponiert → Brute-Force des API-Tokens möglich.
- **Empfehlung:** `httprate` oder ähnliche Middleware einsetzen.

---

## Niedrige Priorität / Code-Qualität

### 13. Interne Error-Details werden exponiert

- **Datei:** `internal/api/handlers.go:524`
- **Problem:** `writeError(..., err.Error())` kann interne Pfade, Konfigurationswerte oder Netzwerkinfos preisgeben.
- **Empfehlung:** In Production-Builds `Details` weglassen; intern loggen.

### 14. Fehlende Nil-Checks

- **Datei:** `internal/api/handlers.go:544, 676`
- **Problem:** `GetMappingManager()` / `GetUDPSender()` werden ohne Nil-Prüfung verwendet.

### 15. Keine Größen-Limits für Mapping-Listen

- **Datei:** `internal/api/handlers.go:568`
- **Problem:** `cfg.Loxone.Mappings` kann unbegrenzt wachsen.
- **Empfehlung:** Max. Mappings (z. B. 1000) erzwingen.

### 16. Mehrfach parallele `RefreshSensorStatus()`-Calls

- **Datei:** `internal/gateway/service.go:321-334`
- **Problem:** Kein Schutz vor gleichzeitigen Aufrufen.
- **Empfehlung:** Mutex oder "Last-Refresh"-Tracking einführen.

### 17. Go-Version 1.21

- **Datei:** `go.mod`
- **Empfehlung:** Auf 1.22 oder 1.23 aktualisieren (aktive Security-Patches, neue `slog`-APIs).

---

## Infrastruktur & Deployment

### Dockerfile

- Non-root User (`appuser`) wird genutzt — gut.
- `config.yaml` sollte read-only gemountet werden (`docker-compose.yml`).
- CMD hardcodiert `-config /app/config.yaml` — akzeptabel, aber via ENV überschreibbar wäre flexibler.

### Dependencies (go.mod)

| Paket | Version | Status |
|-------|---------|--------|
| `github.com/go-chi/chi/v5` | v5.0.12 | aktuell |
| `github.com/gorilla/websocket` | v1.5.1 | aktuell |
| `github.com/rs/zerolog` | v1.32.0 | aktuell |
| `gopkg.in/yaml.v3` | v3.0.1 | aktuell |

**Empfehlung:** Regelmäßig mit `govulncheck ./...` prüfen.

---

## Testing

**Kritischer Befund:** Es gibt keine `*_test.go`-Dateien im gesamten Projekt.

**Empfohlene minimale Testabdeckung:**

- `internal/klf200/commands.go` — Frame-Build / -Parse (Roundtrip-Tests).
- `internal/klf200/client.go` — State-Machine mit Mock-Connection.
- `internal/loxone/mapping.go` — Mapping-Resolution, Edge-Cases.
- `internal/config/config.go` — YAML-Roundtrip, Defaults.
- `internal/api/handlers.go` — Auth, Input-Validierung, Happy-Path.

**Ziel:** Mindestens 60 % Coverage auf den kritischen Paketen.

---

## Web-Frontend

- Vue 3 + Vite + TailwindCSS.
- Keine offensichtlichen XSS-Vektoren gefunden.
- **Empfehlung:** CSRF-Schutz prüfen, falls Cookies/Sessions verwendet werden.

---

## Priorisierte nächste Schritte

| Schritt | Dauer | Priorität |
|---------|-------|-----------|
| 1. Passwort aus Logs und `GetConfig`-Response entfernen | 5 Min | KRITISCH |
| 2. `http.MaxBytesReader` in allen Handlern | 15 Min | KRITISCH |
| 3. TLS-Validierung via `VerifyPeerCertificate` | 30 Min | HOCH |
| 4. Basis-Tests für `klf200/commands.go` und `loxone/mapping.go` | 0,5 Tag | HOCH |
| 5. Context korrekt durch alle Loops propagieren | 1 Std | MITTEL |
| 6. Rate-Limiting-Middleware einbauen | 1 Std | MITTEL |
| 7. UUID-Validierung für Mappings | 30 Min | MITTEL |
| 8. Go-Version auf 1.22/1.23 anheben | 15 Min | NIEDRIG |

---

## Zusammenfassung

| Priorität | Anzahl |
|-----------|--------|
| Kritisch  | 3 |
| Hoch      | 4 |
| Mittel    | 5 |
| Niedrig   | 5 |

**Gesamteinschätzung:** Die Code-Basis ist solide strukturiert, aber die Kombination aus fehlenden Tests, mehreren Credential-Exposures und deaktivierter TLS-Verifikation macht das Projekt aktuell nicht produktionsreif. Mit rund 2–3 Entwicklertagen können alle kritischen und hohen Punkte adressiert werden.
