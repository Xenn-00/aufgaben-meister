# Aufgaben-Meister

Aufgaben-Meister ist ein projectbasiertes Aufgaben- und Rollenmanagement-System, das für klare Verantwortlichkeiten, saubere Audit-Trails und skalierbare Teamarbeit entwickelt wurde.

Der Fokus liegt nicht auf einfachen CRUD, sondern auf Business-Logik, Rollenmodellen und nachvollziehbaren Zustandsänderungen von Aufgaben (Aufgaben).

## 🎯 Ziel des Projekts

Aufgaben-Meister wurde gebaut, um folgende Probleme sauber zu lösen:

- Aufgaben mit klaren Verantwortlichen verwalten
- Rollenbasierte Kontrolle (z.B. Meister vs. Mitarbeiter)
- Jede wichtige Aktion ist nachvollziehbar (Audit Trail)
- Saubere Trennung zwischen HTTP API und asynchroner Verarbeitung
- Produktionsnahe Architektur (Auth, Worker, Queue, Rate Limiting)

Kurz gesagt: ein persönliches Lern- und Architekturprojekt mit realistischen Anforderungen.

## 🧠 Zentrale Konzepte

### Aufgaben (Tasks)

- Aufgaben existieren immer innerhalb eines Projekts
- Eine Aufgaben kann:
  - unassigned sein
  - assigned sein
  - archieviert sein
  - überfällig sein (Overdue = Zustand, kein Endpoint)

### Projekte

- Projekte sind der organisatorische Rahmen
- Benutzer werden über Einladungen Mitgleid
- Rollen bestimmen, was ein Benutzer darf

### Audit Trail

- Jede relevante Aktion erzeugt ein Event
- Keine stillen Zustandsänderungen
- Events sind über API abrufbar

## 🏗️ Architektur - Überblick

Das System besteht aus zwei Hauptkomponenten:

```
[ Client ]
    ↓
    [ HTTP API (Fiber) ] ──→ PostgreSQL
    │
    ├─ Redis (Session, Rate Limit, Cache)
    │
    └─ Async Queue ──→ Worker
```

### 1. HTTP API Server

Verantworlich für:

- Authentifizierung (PASETO + Redis)
- Request-Validierung
- Business-Logik
- Rollenprüfung
- Event-Erzeugung

### 2. Worker Service

Verantwortlich für:

- Asynchrone Tasks (z.B. Emails, Reminder)
- Entkoppelt von HTTP Requests
- Erhöht Stabilität und Skalierbarkeit

## 🔐 Authentifizierung & Sicherheit

- Token-basiert mit PASETO
- Multi-Device Sessions (Redis)
- Logout pro Gerät oder global
- Request-ID & strukturierte Logs
- Rate Limiting für kritische Endpoints

## 📦 Tech Stack

- Go (Backend)
- Fiber (HTTP Framework)
- PostgreSQL (pgx)
- Redis (Session, Cache, Rate Limit)
- PASETO (Token)
- Asynq / Queue-Worker (Async Jobs)

## 📂 Projektstruktur (vereinfacht)

```
cmd/                # Entry points (API & Worker)
internal/
    handlers/       # HTTP Handler (Controller)
    use-cases/      # Business Logik
    repo/           # Datenzugriff (Postgres)
    entity/         # Domain Models
    middleware/     # Auth, Roles, Error Handling
    worker/         # Async Worker
migrations/         # SQL Migrations
```

Regel:
Handler wissen nichts über SQL. Repos wissen nichts über HTTP.

## 🧩 Design-Philosophie

### 1. Klare Intention pro Endpoint

- Kein "do-everything"-Endpoint
- Jeder Endpoint steht für eine Business-Aktion

### 2. Audit > Convenience

- Lieber ein Event mehr als ein stiller State-Change

### 3. Skalierbarkeit vor Bequemlichkeit

- Async Worker statt blockierender HTTP Request

## 🦖 Für wen ist dieses Projekt?

- Backend-Entwickler
- Reviewer / Recruiter
- Teams, die saubere Business-Logik schätzen

PS: Nicht gedacht als Tutorial, sonder als realistische Produktionsarchitektur.

## 🚀 Status

- Core Feature: ✅ fertig
- Aufgaben-Lifecycle: ✅ stabil
- Audit Trail: ✅ vollständig
- Erweiterbar für weitere Domänen

## ❓ Häufige Fragen (FAQ)

### Ist dieses Projekt produktiv im Einsatz?

Nein. Der Schwerpunkt lag nicht auf Deployment, sondern auf Architektur und sauberen Entscheidungen. Das Projekt dient dazu zu zeigen, wie Aufgaben, Rollen und Zustandsänderungen in einem realistischen Backend modelliert werden können.

### Warum bezeichnest du das Projekt nicht als reines Demo- oder Spielzeugprojekt?

Weil bewusst Konzepte umgesetzt wurden, die auch in realen Systemen relevant sind: Rollenmodelle, Audit Trails, klare Zustandsübergänge, Trennung von API und asynchroner Verarbeitung.

### Gibt es Monitoring, SLA oder produktive Nutzer?

Nein. Diese Aspekte waren nicht Ziel des Projekts. Der Fokus lag auf Backend-Design, Wartbarkeit und Nachvollziehbarkeit der Business-Logik.

### Für wen ist dieses Projekt gedacht?

Für Reviewer, Ausbilder und Entwickler, die verstehen möchten, wie ich an Backend-Architektur und fachliche Probleme herangehe. Nicht als fertiges kommerzielles Produkt.
