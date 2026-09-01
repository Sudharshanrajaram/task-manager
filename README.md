# TaskFlow

> High-performance personal multi-project task manager with concurrent timers, real-time sync, Obsidian-style notes, daily logs & Excel export, and self-improving RAG subtask suggestions.

TaskFlow is built with **Go (Gin)** and **React (TypeScript + Tailwind CSS)** for engineers managing parallel projects, distraction-free focus sessions, and AI-accelerated workflows.

---

## Architecture Overview

- **Backend**: Go 1.24, Gin Web Framework, GORM (PostgreSQL / SQLite fallback), Gorilla WebSockets, Redis (Pub/Sub & Cache)
- **Database**: PostgreSQL 16 + `pgvector` extension (with automatic SQLite fallback for zero-dependency local dev)
- **AI Engine**: Groq API (`openai/gpt-oss-20b` / `groq/compound-mini` / `qwen3.6-27b`)
- **Frontend**: React 18, TypeScript, Tailwind CSS, TanStack Query, Zustand, Lucide Icons, cmdk
- **Containerization & CI**: Docker multi-stage builds, Docker Compose full-stack orchestration, GitHub Actions CI

---

## Quick Start

### Option A: Complete Full-Stack via Docker Compose (Recommended)

Start Postgres (with `pgvector`), Redis, Go API, and React Frontend in one command:

```bash
docker compose up --build -d
```

- **Frontend Web UI**: `http://localhost:3000`
- **Backend API**: `http://localhost:8080`
- **Health Check**: `http://localhost:8080/health`
- **Prometheus Metrics**: `http://localhost:8080/metrics`

---

### Option B: Local Development

#### 1. Start Database & Redis (Optional but recommended for pgvector & WebSockets)
```bash
make db-up
```
*(Note: If Docker is not running, TaskFlow will automatically fall back to local SQLite `backend/taskflow_dev.db` and an in-memory WebSocket event bridge).*

#### 2. Start Both Backend & Frontend
```bash
make dev
```
- Frontend will be available at `http://localhost:5173`
- Backend will be available at `http://localhost:8080`

#### 3. Run Test Suite
```bash
make test
```

---

## Configuration (`backend/.env`)

| Variable | Description | Default |
|---|---|---|
| `PORT` | API Server Port | `8080` |
| `ENV` | Environment (`development` or `production`) | `development` |
| `DB_HOST` | Postgres Host | `localhost` |
| `DB_PORT` | Postgres Port | `5432` |
| `DB_USER` | Postgres User | `taskflow` |
| `DB_PASSWORD` | Postgres Password | `taskflow_secret` |
| `DB_NAME` | Postgres Database Name | `taskflow_db` |
| `REDIS_HOST` | Redis Host | `localhost` |
| `REDIS_PORT` | Redis Port | `6379` |
| `JWT_SECRET` | Secret key for signing JWT tokens | `min-32-chars-secret` |
| `GROQ_API_KEY` | Groq API Key (get free at https://console.groq.com) | `gsk_...` |
| `GROQ_CHAT_MODEL` | Groq LLM Model | `openai/gpt-oss-20b` |

---

## Completed Roadmap Phases

- [x] **Phase 1**: Project Setup & Clean Backend Skeleton
- [x] **Phase 2**: Core Ticket/Task CRUD & Migrations
- [x] **Phase 3**: Concurrency: Timer Engine (Goroutines, Channels, Race-Safe State)
- [x] **Phase 4**: RAG: AI Subtask Suggestions (pgvector + Groq Chat & Deterministic Vectors)
- [x] **Phase 5**: Frontend Build (React + Tailwind + Design System + Command Palette)
- [x] **Phase 6**: Realtime Sync (WebSockets + Redis Pub/Sub & Fallback Event Bridge)
- [x] **Phase 7**: JWT Auth & User Isolation (Access/Refresh rotation + Zustand AuthStore)
- [x] **Phase 8**: Testing & Observability (Go race test suite, Prometheus `/metrics`, Deep `/health`)
- [x] **Phase 9**: Deployment & CI/CD (Multi-stage Dockerfiles, Docker Compose, GitHub Actions CI)
- [x] **Phase 10**: Daily Log Dashboard, Auto-Archiving & Native Excel `.xlsx` Export
- [x] **Phase 11**: Focus Mode Notes (Obsidian-Style Markdown, Debounced Autosave, `[[backlinks]]`)
- [x] **Phase 12**: Task Summary "Ask" Bar (Content-hash SHA256 cached LLM summaries)
- [x] **Phase 13**: Stretch Goals (AI Daily Standup Generator + Task Dependency Graph)
