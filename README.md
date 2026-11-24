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

## Quick Start

The fastest way to get the full stack running — PostgreSQL + pgvector, Ollama (containerized LLM), Go backend, and React frontend — is with a single command:

```bash
# Clone the repo
git clone https://github.com/akhmadzaki/go-backend.git
cd go-backend

# Copy environment file
cp .env.example .env

# Start all services (PostgreSQL, Ollama, backend, frontend)
make run

# Open the frontend
open http://localhost:3000
```

This starts four containers:
| Service | Image | Port |
|---------|-------|------|
| [`postgres`](docker-compose.yml#L104) | `pgvector/pgvector:pg16` | `5432` |
| [`ollama`](docker-compose.yml#L36) | `ollama/ollama` | `11434` |
| [`backend`](docker-compose.yml#L129) | Built from [`Dockerfile`](Dockerfile) | `8080` |
| [`frontend`](docker-compose.yml#L232) | Built from [`frontend/Dockerfile`](frontend/Dockerfile) | `3000` |

> **No manual Ollama installation required.** The LLM runs inside a Docker container. Models are pulled automatically on first run via `make run`.

### GPU Acceleration

If you have an **NVIDIA GPU** with the [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html) installed:

```bash
# Production with NVIDIA GPU
make run-gpu
```

On **Apple Silicon (MPS / Metal)**, Ollama uses GPU acceleration automatically — no special Docker configuration is needed:

```bash
OLLAMA_PROFILE=mps make run
```

The GPU backend is selected via Docker Compose **profiles**. See the [`ollama` services](docker-compose.yml#L21) for details.

To stop all services:
```bash
make stop
```

To stop and delete all data (database + downloaded models):
```bash
make stop   # then
docker compose down -v
```

---

## OS-Specific Notes

Docker handles cross-platform compatibility, but each OS has unique considerations for GPU acceleration:

| OS | GPU Support | Notes |
|----|-------------|-------|
| **Linux** | NVIDIA GPU via [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html) | Use `make run-gpu` (activates [`ollama-gpu`](docker-compose.yml#L57) with `--profile nvidia`) |
| **macOS (Apple Silicon)** | Metal GPU (automatic via Rosetta 2) | Use `OLLAMA_PROFILE=mps make run` (activates [`ollama-mps`](docker-compose.yml#L86) with `--profile mps`) |
| **macOS (Intel)** | CPU only | `make run` — no changes required |
| **Windows** | NVIDIA GPU via [NVIDIA Container Toolkit for WSL2](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html) | Use `make run-gpu` (activates [`ollama-gpu`](docker-compose.yml#L57) with `--profile nvidia`) |

> **Prerequisites:** [Docker](https://docs.docker.com/engine/install/), [Git](https://git-scm.com/), and [GNU Make](https://www.gnu.org/software/make/) are required on all platforms.

---

## Local Development

Run the backend and frontend directly on your machine while PostgreSQL and Ollama run in Docker.

### 1. Start Dependencies

```bash
# Start PostgreSQL + pgvector and Ollama (CPU)
docker compose --profile cpu up -d postgres ollama

# Or with NVIDIA GPU:
# docker compose --profile nvidia up -d postgres ollama-gpu
```

### 2. Configure Environment

```bash
cp .env.example .env
```

Key environment variables (defaults work for local development):

| Variable | Default | Notes |
|----------|---------|-------|
| `APP_PORT` | `8080` | Backend API server port |
| `DB_HOST` | `localhost` | `postgres` inside Docker Compose network |
| `EMBEDDING_PROVIDER` | `ollama` | `ollama` or `openai` |
| `EMBEDDING_MODEL` | `nomic-embed-text` | Must be pulled in Ollama |
| `LLM_PROVIDER` | `local` | `local` (Ollama), `openai`, or `anthropic` |
| `LLM_MODEL` | `llama3.2:1b-instruct-q2_K` | Must be pulled in Ollama |
| `CHUNK_STRATEGY` | `sentence` | `fixed`, `sentence`, or `paragraph` |

### 3. Run the Backend

```bash
# Install dependencies
go mod tidy

# Run database migrations (automatic on server start when DB_AUTO_MIGRATE=true)
# Or manually:
go run ./cmd/migrate up

# Start the server
go run main.go
```

Verify: `curl http://localhost:8080/api/v1/health` → `{"status":"ok"}`

### 4. Run the Frontend

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

The Vite dev server starts on `http://localhost:5173`.

### 5. Verify End-to-End

1. Open the frontend at `http://localhost:5173` (dev) or `http://localhost:3000` (Docker)
2. Upload a document (PDF, Markdown, HTML, or plain text)
3. Search your documents using hybrid search
4. Start a conversation in the chat interface

---

## Interactive API Documentation (Swagger)

The backend serves Swagger UI with auto-generated OpenAPI 2.0 documentation:

| Endpoint | Description |
|----------|-------------|
| `GET http://localhost:8080/swagger/index.html` | Interactive Swagger UI |
| `GET http://localhost:8080/swagger/doc.json` | Raw OpenAPI spec (JSON) |

Regenerate after changing endpoint annotations:
```bash
make swagger
```

---

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| `pq: role "postgres" does not exist` | PostgreSQL user not created | `docker compose down -v && docker compose up -d postgres` |
| `pq: extension "vector" not available` | Wrong PostgreSQL image | Use `pgvector/pgvector:pg16` (not plain `postgres:16`) |
| `connection refused` on `:5432` | PostgreSQL not running | `docker compose up -d postgres` |
| `connection refused` on `:11434` | Ollama not running or still starting | `docker compose ps` — check for `ollama`, `ollama-gpu`, or `ollama-mps` |
| `model "llama3.2:1b-instruct-q2_K" not found` | Model not pulled | `docker compose exec ollama ollama pull llama3.2:1b-instruct-q2_K` (or `ollama-gpu` for GPU) |
| Ollama container exits immediately | Insufficient memory or port conflict | Increase Docker memory to 4 GB+; ensure port `11434` is free |
| Frontend shows blank page / CORS errors | Backend not running | Ensure `go run main.go` or `docker compose up -d` is running |
| `go: module ... not found` | Dependencies not downloaded | Run `go mod tidy` |
| Uploaded document stuck on `pending` | Embedding model not pulled | `docker compose exec ollama ollama pull nomic-embed-text` (or `ollama-gpu` for GPU) |
| `nvidia-container-toolkit` errors on `make run-gpu` | NVIDIA Container Toolkit not installed | Install from [NVIDIA docs](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html) or use `make run` (CPU fallback) |

---

## Next Steps

- **Switch to OpenAI embeddings** — Set `EMBEDDING_PROVIDER=openai` and add `OPENAI_API_KEY=sk-...` to your `.env`
- **Use a remote LLM** — Set `LLM_PROVIDER=openai` and configure `LLM_MODEL=gpt-4o-mini` with your API key
- **Tune chunking** — Experiment with `CHUNK_STRATEGY` (`fixed`, `sentence`, `paragraph`) and `CHUNK_SIZE`
- **Enable LlamaIndex sidecar** — Deploy the LlamaIndex Python service for advanced indexing strategies (see [`knowledge/index/llama_index.go`](knowledge/index/llama_index.go))
- **Production deployment** — Set `APP_ENV=production`, `APP_DEBUG=false`, and use a process manager or container orchestration

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