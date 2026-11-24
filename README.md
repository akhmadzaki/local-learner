# Local AI Knowledge Base / RAG System

A full-stack AI-powered knowledge base for personal learning. Upload documents, build a local knowledge graph with RAG (Retrieval-Augmented Generation), and chat with an AI model — all running locally on your machine.

## Features

- **Document Ingestion** — Upload PDF, Markdown, HTML, or plain text files
- **Intelligent Chunking** — Split documents into searchable pieces (fixed, sentence, or paragraph strategies)
- **Vector Embeddings** — Generate embeddings using local or OpenAI models
- **Hybrid Search** — Combine vector similarity and keyword (full-text) search
- **AI Chat** — Conversational agent with context-aware answers using your documents
- **Local-First** — Designed to work with local LLMs (Ollama, llama.cpp, etc.) and local embedding models
- **Machine-Aware** — Configure for CPU, CUDA (GPU), or MPS (Apple Silicon) based on your hardware

## Architecture

```
┌─────────────┐     ┌─────────────────────────────────────┐
│  Frontend   │     │            Backend (Go)              │
│  React/TS   │◄───►│  ┌───────────────────────────────┐  │
│             │     │  │  API Layer (net/http)         │  │
│  Chat UI    │     │  │  Routes → Controllers         │  │
│  Search UI  │     │  ├───────────────────────────────┤  │
│  Upload UI  │     │  │  Services                     │  │
│             │     │  │  Document → Chunk → Embed     │  │
│             │     │  │  Search → Agent → Chat        │  │
└─────────────┘     │  ├───────────────────────────────┤  │
                    │  │  Vector Store (pgvector)      │  │
                    │  │  PostgreSQL + pgvector         │  │
                    │  └───────────────────────────────┘  │
                    └─────────────────────────────────────┘
```

## Quick Start (Docker)

The fastest way to get the full stack running — PostgreSQL + pgvector, Go backend, and React frontend — is with Docker Compose:

```bash
# Clone the repo
git clone https://github.com/akhmadzaki/go-backend.git
cd go-backend

# Copy environment file
cp .env.example .env

# Start all services (PostgreSQL + pgvector, backend, frontend)
docker compose up -d

# Open the frontend
open http://localhost:3000
```

This starts three containers:
| Service | Image | Port |
|---------|-------|------|
| [`postgres`](docker-compose.yml#L23) | `pgvector/pgvector:pg16` | `5432` |
| [`backend`](docker-compose.yml#L42) | Built from [`Dockerfile`](Dockerfile) | `8080` |
| [`frontend`](docker-compose.yml#L82) | Built from [`frontend/Dockerfile`](frontend/Dockerfile) | `3000` |

> The PostgreSQL container includes the pgvector extension out of the box — no manual installation needed. The database `go_backend`, user `postgres`, and the `vector` extension are all configured automatically via the [`docker-compose.yml`](docker-compose.yml#L29-L34) environment variables.

To stop: `docker compose down`
To stop and delete data: `docker compose down -v`

---

## Local Setup Tutorial

This guide walks you through running each component locally for development. PostgreSQL + pgvector runs via Docker (no manual install needed), while the backend and frontend run directly on your machine for hot-reload and debugging.

---

### 1. Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| [Go](https://go.dev/dl/) | 1.21+ | Backend server |
| [Node.js](https://nodejs.org/) | 18+ | Frontend build tooling |
| [Docker](https://docs.docker.com/engine/install/) | 24+ | PostgreSQL + pgvector container |
| [Ollama](https://ollama.com/download) | Latest | Local LLM inference |

> **Hardware note:** This project is designed to be machine-aware. Set [`EMBEDDING_DEVICE`](#environment-variables) to `cpu`, `cuda` (NVIDIA GPU), or `mps` (Apple Silicon) depending on your hardware.

---

### 2. Database Setup (PostgreSQL + pgvector via Docker)

PostgreSQL with the pgvector extension runs in a Docker container. No manual PostgreSQL or pgvector installation is required.

#### 2.1 Start the Database

```bash
# Start only the PostgreSQL + pgvector container
docker compose up -d postgres
```

This uses the [`pgvector/pgvector:pg16`](https://hub.docker.com/r/pgvector/pgvector) image ([`docker-compose.yml`](docker-compose.yml#L23-L39)), which comes with pgvector pre-installed. The container automatically:
- Creates the `go_backend` database
- Creates the `postgres` user
- Enables the `vector` extension

#### 2.2 Verify the Database

```bash
# Check that PostgreSQL is healthy
docker compose ps postgres

# Connect to the database and verify pgvector
docker compose exec postgres psql -U postgres -d go_backend -c "\dx"
```

You should see the `vector` extension listed.

> **Password:** The default password is `postgres` (set via [`POSTGRES_PASSWORD`](docker-compose.yml#L32)). To change it, set `DB_PASSWORD` in your `.env` file before running `docker compose up`.

---

### 3. Ollama Setup (Local LLM)

The backend uses Ollama to serve a local LLM for the AI chat agent. The default model is `llama-3.2-3b-instruct`.

```bash
# Install Ollama (Linux / macOS)
curl -fsSL https://ollama.com/install.sh | sh

# Start the Ollama service
ollama serve
```

In a **separate terminal**, pull the default model:

```bash
# Pull the default model (3B parameters — runs on most machines)
ollama pull llama-3.2-3b-instruct

# (Optional) Pull a smaller model for low-resource machines
ollama pull llama-3.2-1b-instruct

# (Optional) Pull a larger model if you have sufficient RAM/VRAM
ollama pull llama-3.2-8b-instruct
```

> **Verify Ollama is running:** Open `http://localhost:11434` in your browser. You should see `Ollama is running`.

To use a different model, update [`LLM_MODEL`](#environment-variables) in your `.env` file.

---

### 4. Backend Setup

#### 4.1 Clone and Enter the Project

```bash
git clone https://github.com/akhmadzaki/go-backend.git
cd go-backend
```

#### 4.2 Configure Environment

```bash
# Copy the example environment file
cp .env.example .env
```

Open [`.env`](.env) and adjust values as needed. The defaults work out-of-the-box for a local setup:

| Variable | Default | Notes |
|----------|---------|-------|
| `APP_PORT` | `8080` | Backend API server port |
| `DB_HOST` | `localhost` | Use `localhost` for Docker on same host; `postgres` inside Docker Compose network |
| `DB_PORT` | `5432` | Default PostgreSQL port |
| `DB_NAME` | `go_backend` | Created automatically by the Docker container |
| `DB_USER` | `postgres` | |
| `DB_PASSWORD` | `postgres` | Must match [`POSTGRES_PASSWORD`](docker-compose.yml#L32) in docker-compose.yml |
| `EMBEDDING_PROVIDER` | `local` | `local` (stub) or `openai` (requires API key) |
| `EMBEDDING_MODEL` | `all-MiniLM-L6-v2` | Embedding model name |
| `EMBEDDING_DIMENSION` | `384` | Must match the model's output dimension |
| `EMBEDDING_DEVICE` | `cpu` | `cpu`, `cuda`, or `mps` |
| `LLM_PROVIDER` | `local` | `local` (Ollama), `openai`, or `anthropic` |
| `LLM_MODEL` | `llama-3.2-3b-instruct` | Must be pulled in Ollama (see [Step 3](#3-ollama-setup-local-llm)) |
| `LLM_BASE_URL` | `http://localhost:11434` | Ollama API endpoint |
| `CHUNK_STRATEGY` | `sentence` | `fixed`, `sentence`, or `paragraph` |
| `CHUNK_SIZE` | `512` | Max tokens per chunk |
| `CHUNK_OVERLAP` | `64` | Overlap tokens between chunks |

> **Docker Compose note:** When running the backend via [`docker compose`](#quick-start-docker), the environment variables are overridden by [`docker-compose.yml`](docker-compose.yml#L50-L76). Notably, [`DB_HOST`](docker-compose.yml#L55) is set to `postgres` (the service name) so the backend can reach the database container over the internal network.

#### 4.3 Install Go Dependencies

```bash
go mod tidy
```

This downloads the project's dependencies: [`lib/pq`](https://github.com/lib/pq) (PostgreSQL driver), [`golang-migrate`](https://github.com/golang-migrate/migrate) (schema migrations), and their transitive dependencies.

#### 4.4 Run Database Migrations

Migrations use [golang-migrate](https://github.com/golang-migrate/migrate) with SQL files embedded into the binary via `//go:embed`. They run **automatically** on server start (controlled by [`DB_AUTO_MIGRATE`](#environment-variables), default `true`).

The following tables are created:

| Migration | Table | Purpose |
|-----------|-------|---------|
| [`000001_create_documents`](db/migrations/000001_create_documents.up.sql) | `documents` | Uploaded document metadata |
| [`000002_create_chunks`](db/migrations/000002_create_chunks.up.sql) | `chunks` | Document chunks with full-text search (`TSVECTOR`) |
| [`000003_create_embeddings`](db/migrations/000003_create_embeddings.up.sql) | `embeddings` | Vector embeddings (`VECTOR(384)`) with IVFFlat index |
| [`000004_create_conversations`](db/migrations/000004_create_conversations.up.sql) | `conversations`, `conversation_turns` | Multi-turn chat history |

**Manual migration commands** (via the embedded [`cmd/migrate`](cmd/migrate/main.go) CLI):

```bash
# Apply all pending migrations
go run ./cmd/migrate up

# Roll back the most recent migration
go run ./cmd/migrate down

# Migrate to a specific version (up or down)
go run ./cmd/migrate to 2

# Show the current schema version
go run ./cmd/migrate version

# Drop all managed tables (destructive!)
go run ./cmd/migrate drop
```

Or use the Makefile shortcuts:

```bash
make migrate-up
make migrate-down
make migrate-version
make migrate-to VERSION=2
make migrate-create NAME=add_user_preferences
```

To run migrations without starting the HTTP server:

```bash
go run main.go -migrate-only
```

#### 4.5 Start the Backend Server

```bash
go run main.go
```

You should see output similar to:

```
Starting 1.0.0 server on port 8080 (env: development)
db.Connect: connected to PostgreSQL at localhost:5432/go_backend
Server listening on :8080
```

> **Verify the server is running:**
> ```bash
> curl http://localhost:8080/api/v1/health
> ```
> Expected response: `{"status":"ok"}` (or similar health-check payload).

---

### 5. Frontend Setup

#### 5.1 Configure Frontend Environment

```bash
cd frontend
cp .env.example .env
```

The default [`frontend/.env`](frontend/.env) points the Vite dev server to the backend at `http://localhost:8080/api/v1`. Change [`VITE_API_BASE_URL`](frontend/.env.example) if your backend runs on a different host or port.

#### 5.2 Install Dependencies

```bash
npm install
```

#### 5.3 Start the Dev Server

```bash
npm run dev
```

The Vite dev server starts on `http://localhost:5173` by default. Open this URL in your browser.

> **Verify the frontend connects to the backend:** Open the browser's developer console (F12). You should see network requests to `http://localhost:8080/api/v1/...` succeeding.

---

### 6. Verify the Full Stack

Once all services are running, confirm the system works end-to-end:

1. **Open the frontend** — At `http://localhost:5173` (Vite dev server) or `http://localhost:3000` (Docker Compose)
2. **Upload a document** — Use the upload UI to add a PDF, Markdown, HTML, or plain text file
3. **Check the document status** — It should move from `pending` → `processing` → `completed`
4. **Search your documents** — Use the hybrid search bar to find content
5. **Start a conversation** — Ask a question about your uploaded documents in the chat interface

---

### 7. Interactive API Documentation (Swagger)

The backend serves Swagger UI with auto-generated OpenAPI 2.0 documentation for every endpoint. This is useful for exploring the API, trying out requests, and understanding request/response schemas without reading code.

| Endpoint | Description |
|----------|-------------|
| `GET http://localhost:8080/swagger/index.html` | Interactive Swagger UI (browse in a browser) |
| `GET http://localhost:8080/swagger/doc.json` | Raw OpenAPI spec (JSON) |

Open the Swagger UI in your browser:

```bash
open http://localhost:8080/swagger/index.html
```

> The spec is generated from annotations in [`api/controllers/knowledge_controller.go`](api/controllers/knowledge_controller.go) and [`main.go`](main.go) by the [`swag`](https://github.com/swaggo/swag) CLI.

#### Regenerating the docs

Install the CLI once (requires Go):

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

Regenerate after changing any endpoint annotations:

```bash
make swagger
```

This runs `swag init` against [`main.go`](main.go) and rewrites the generated files under [`docs/`](docs/).

---

### 8. Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| `pq: role "postgres" does not exist` | PostgreSQL user not created by Docker | Run `docker compose down -v && docker compose up -d postgres` to recreate |
| `pq: database "go_backend" does not exist` | Database not created by Docker | Check `POSTGRES_DB` in [`docker-compose.yml`](docker-compose.yml#L30) — should be `go_backend` |
| `pq: extension "vector" not available` | pgvector not available in image | Ensure you're using `pgvector/pgvector:pg16` (not plain `postgres:16`) |
| `connection refused` on `:5432` | PostgreSQL container not running | Run `docker compose up -d postgres` and check with `docker compose ps` |
| `connection refused` on `:11434` | Ollama not running | Run `ollama serve` in a separate terminal |
| `model "llama-3.2-3b-instruct" not found` | Model not pulled in Ollama | Run `ollama pull llama-3.2-3b-instruct` |
| Frontend shows blank page / CORS errors | Backend not running or wrong port | Ensure `go run main.go` or `docker compose up -d` is running |
| `go: module ... not found` | Dependencies not downloaded | Run `go mod tidy` |
| Uploaded document stuck on `pending` | Embedding provider not configured | The `local` embedding provider is a stub — switch to `openai` or implement a real local provider |

---

### 9. Next Steps

- **Switch to OpenAI embeddings** — Set `EMBEDDING_PROVIDER=openai` and add `OPENAI_API_KEY=sk-...` to your `.env`
- **Use a remote LLM** — Set `LLM_PROVIDER=openai` and configure `LLM_MODEL=gpt-4o-mini` with your API key
- **Tune chunking** — Experiment with [`CHUNK_STRATEGY`](#environment-variables) (`fixed`, `sentence`, `paragraph`) and [`CHUNK_SIZE`](#environment-variables) for your document types
- **Enable LlamaIndex sidecar** — Deploy the LlamaIndex Python service for advanced indexing strategies (see [`knowledge/index/llama_index.go`](knowledge/index/llama_index.go))
- **Production deployment** — Set `APP_ENV=production`, `APP_DEBUG=false`, and use a process manager like `supervisord` or a container orchestration platform

## Project Structure

```
├── main.go                    # Entry point
├── cmd/
│   └── migrate/               # golang-migrate CLI (up/down/version/…)
├── config/                    # Application configuration
├── api/
│   ├── controllers/           # HTTP handlers
│   └── routes/                # Route definitions
├── db/
│   ├── db.go                  # PostgreSQL connection
│   ├── migrate.go             # golang-migrate runner (embedded SQL)
│   ├── migrations/            # SQL migration files (up/down)
│   ├── models/                # Database models
│   └── repository/            # Data access layer
├── dto/                       # Data transfer objects
├── helpers/                   # HTTP response helpers
├── knowledge/
│   ├── agent/                 # AI chat agent & reasoning
│   ├── chunks/                # Document chunking strategies
│   ├── documents/             # Document processing & ingestion
│   ├── embeddings/            # Embedding providers & service
│   ├── index/                 # Search index management
│   ├── pipeline/              # RAG pipeline orchestration
│   ├── search/                # Hybrid, vector, keyword search
│   └── vectorstore/           # pgvector integration
├── logger/                    # Structured logging
├── types/                     # Shared domain types
└── frontend/                  # React TypeScript frontend
    ├── src/
    │   ├── api/               # API client
    │   ├── components/        # React components
    │   ├── hooks/             # React hooks
    │   ├── types/             # TypeScript types
    │   └── utils/             # Utility functions
    └── ...