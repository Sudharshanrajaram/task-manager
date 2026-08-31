# TaskFlow

> Personal Multi-Project Task Manager with Concurrent Timers & RAG-Assisted Subtasks.

TaskFlow is a high-performance productivity tool built in **Go (Gin)** and **React (TypeScript + Tailwind)** designed for juggling concurrent projects, nested task/subtask timers, and self-improving subtask generation using pgvector and Groq.

---

## Architecture Overview

- **Backend**: Go 1.22+, Gin Web Framework, GORM / pgvector, Gorilla WebSockets, Redis (Pub/Sub & Cache)
- **Database**: PostgreSQL 16 + `pgvector` extension
- **AI / Embeddings**: Groq API (`llama-3.1-8b-instant` / `llama-3.3-70b-versatile` + `nomic-embed-text-v1_5`)
- **Frontend**: React 18, TypeScript, Tailwind CSS, Zustand / TanStack Query, cmdk

---

## Quick Start (Backend)

### 1. Prerequisites
- [Go 1.22+](https://go.dev/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) or [OrbStack](https://orbstack.dev/) (for Postgres + pgvector & Redis)

### 2. Start Infrastructure
```bash
cd backend
docker-compose up -d
```

### 3. Configure Environment
```bash
cp .env.example .env
```

### 4. Run the API Server
```bash
cd backend
go run cmd/api/main.go
```
Verify the server is running:
```bash
curl http://localhost:8080/health
```

---

## Roadmap & Phases

- [x] **Phase 1**: Project Setup & Backend Skeleton
- [ ] **Phase 2**: Core Ticket/Task CRUD & Migrations
- [ ] **Phase 3**: Concurrency: The Timer Engine (Goroutines, Channels, Race-safe state)
- [ ] **Phase 4**: RAG: AI Subtask Suggestions (pgvector + Groq Embeddings/Chat)
- [ ] **Phase 5**: Frontend Build (React + Tailwind + Design System)
- [ ] **Phase 6**: Realtime Sync (WebSockets + Redis Pub/Sub)
- [ ] **Phase 7**: JWT Auth & Data Isolation
- [ ] **Phase 8**: Testing & Observability
- [ ] **Phase 9**: Deployment & CI/CD

