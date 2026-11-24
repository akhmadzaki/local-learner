# Production Readiness & Portfolio Review

> Comprehensive audit of the Local AI Knowledge Base / RAG System project.
> Generated: 2026-11-23

---

## Overview

The project has a **solid foundation** — clean architecture, Docker setup, structured logging, database migrations, and a well-organized RAG pipeline. Below is a prioritized roadmap to make it production-ready and standout as a portfolio project.

---

## 🔴 Critical (Security & Stability)

### 1. SQL Injection — Table names concatenated without sanitization

**Where:** [`knowledge/vectorstore/vector_store.go`](knowledge/vectorstore/vector_store.go)

Lines 37, 49, 74, and 85 build SQL queries by concatenating `s.TableName` directly into the query string. While the table name is set internally (not from user input), this is still a code smell and prevents static analysis tools from verifying safety.

**Fix:** Whitelist the table name or use a constant. Example:

```go
func (s *VectorStore) UpsertVector(...) bool {
    query := `INSERT INTO embeddings (chunk_id, embedding, metadata) VALUES ($1, $2, $3)
              ON CONFLICT (chunk_id) DO UPDATE SET embedding = $2, metadata = $3`
    _, err := s.db.Exec(query, chunkID, vector, metadata)
    ...
}
```

### 2. No graceful shutdown

**Where:** [`main.go`](main.go:34)

`http.ListenAndServe` blocks indefinitely. A SIGTERM/SIGINT kills the process immediately, dropping in-flight requests and corrupting state.

**Fix:** Use `signal.NotifyContext` to catch OS signals and gracefully drain the server:

```go
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()

srv := &http.Server{Addr: addr, Handler: router}
go func() {
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("Server error: %v", err)
    }
}()

<-ctx.Done()
log.Println("Shutting down gracefully...")
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
srv.Shutdown(shutdownCtx)
```

### 3. No HTTP timeouts

**Where:** [`main.go`](main.go:34)

The server has no `ReadTimeout`, `WriteTimeout`, or `IdleTimeout`, making it vulnerable to slow-loris attacks and resource exhaustion.

**Fix:**

```go
srv := &http.Server{
    Addr:         addr,
    Handler:      router,
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 60 * time.Second, // LLM requests can be slow
    IdleTimeout:  60 * time.Second,
}
```

### 4. No CORS middleware

**Where:** `main.go` / routes

The frontend runs on a different origin (e.g., `localhost:5173` in dev, `localhost:3000` in Docker). Without CORS headers, browsers block cross-origin requests in production.

**Fix:** Add a CORS middleware wrapping the mux:

```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

> **Note:** The Nginx config in Docker already proxies `/api/` to the backend, so CORS is less critical in the Docker deployment. But for local development (frontend on Vite, backend on Go), it's needed.

### 5. No rate limiting

**Where:** `main.go`

The API is unprotected against abuse. A malicious actor could flood the embedding endpoint or chat endpoint, incurring costs (if using OpenAI) or degrading performance.

**Fix:** Add a simple token-bucket rate limiter middleware or use a library like `golang.org/x/time/rate`.

---

## 🟡 High (Portfolio Impact)

### 6. Zero tests

**Where:** Entire project

There are no `_test.go` files anywhere. This is the single biggest gap for both production-readiness and portfolio presentation.

**Fix:** Add tests for:

| Package | Priority | What to test |
|---------|----------|-------------|
| `knowledge/documents` | 🔴 | `IngestDocument`, `DeleteDocument` |
| `knowledge/embeddings` | 🔴 | `OpenAIEmbeddingProvider.CreateEmbedding` (mock HTTP) |
| `knowledge/search` | 🔴 | `Search` dispatch logic, `toSearchResponse` |
| `knowledge/chunks` | 🟡 | Each chunk strategy (fixed, sentence, paragraph) |
| `knowledge/vectorstore` | 🟡 | `UpsertVector`, `Search`, `DeleteByDocument` (with test DB) |
| `api/controllers` | 🟡 | HTTP handler tests with `httptest.NewRecorder` |
| `db/repository` | 🟡 | Repository CRUD with testcontainers or in-memory |
| `frontend/src/` | 🟡 | Component rendering with Vitest + React Testing Library |

**Example test structure:**

```
db/repository/
  document_repository_test.go
  chunk_repository_test.go
knowledge/documents/
  document_service_test.go
knowledge/embeddings/
  embedding_providers_test.go
knowledge/search/
  search_service_test.go
api/controllers/
  knowledge_controller_test.go
```

### 7. No CI/CD

**Where:** Project root

No GitHub Actions, GitLab CI, or similar pipeline exists. This is a red flag for production-readiness.

**Fix:** Add `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.21"
      - run: go mod tidy
      - run: go vet ./...
      - run: go test ./... -v -race -count=1

  frontend:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: frontend
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: "20"
      - run: npm ci
      - run: npm run lint
      - run: npm run build

  docker:
    needs: [backend, frontend]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: docker compose build
```

### 8. No context propagation

**Where:** All service layers

`context.Context` is not threaded through any service or repository call. This means:
- HTTP request cancellation doesn't propagate to DB queries
- Deadlines/timeouts can't be set per-request
- Tracing/metadata can't flow through the call chain

**Fix:** Thread `context.Context` from the HTTP handler through every service and repository method:

```go
// Handler
func (ctrl *KnowledgeController) GetDocument(w http.ResponseWriter, r *http.Request) {
    docID := r.PathValue("id")
    doc := ctrl.docSvc.GetDocument(r.Context(), docID)  // pass context
    ...
}

// Service
func (svc *DocumentService) GetDocument(ctx context.Context, docID string) *models.Document {
    return svc.docRepo.FindByID(ctx, docID)
}

// Repository
func (r *DocumentRepository) FindByID(ctx context.Context, id string) *models.Document {
    row := r.db.QueryRowContext(ctx, "SELECT ...", id)
    ...
}
```

### 9. Error handling is inconsistent

**Where:** Multiple files

The codebase mixes:
- `log.Printf` for errors
- `helpers.RespondError` for HTTP responses
- Bare `return nil, err` without wrapping
- The `errors/` package exists but is **empty**

**Fix:** Implement typed domain errors in the `errors/` package:

```go
package errors

type AppError struct {
    Code       string `json:"code"`
    Message    string `json:"message"`
    HTTPStatus int    `json:"-"`
    Err        error  `json:"-"`
}

var (
    ErrDocumentNotFound = &AppError{Code: "DOCUMENT_NOT_FOUND", HTTPStatus: 404}
    ErrIngestionFailed  = &AppError{Code: "INGESTION_FAILED", HTTPStatus: 500}
    ErrInvalidRequest   = &AppError{Code: "INVALID_REQUEST", HTTPStatus: 400}
)
```

Then create a centralized error handler middleware that converts `AppError` to JSON responses.

### 10. Local embedding provider is a stub

**Where:** [`knowledge/embeddings/embedding_providers.go`](knowledge/embeddings/embedding_providers.go:74)

The `LocalEmbeddingProvider.CreateEmbedding` method returns a zero vector and logs a warning. This means the "local-first" promise is broken — users must use OpenAI for real embeddings.

**Fix (options):**

| Option | Effort | Quality |
|--------|--------|---------|
| **A.** Integrate `github.com/pdevine/tensor` or `github.com/nicholasgasior/gosbert` for local embeddings | High | Best |
| **B.** Call Ollama's embedding endpoint (`/api/embeddings`) which supports `nomic-embed-text` | Medium | Good |
| **C.** Document the limitation clearly and add a prominent warning in the README | Low | Acceptable |

**Recommended:** Option B — Ollama already runs as a sidecar and supports embeddings natively:

```go
type OllamaEmbeddingProvider struct {
    BaseURL string
    Model   string // e.g., "nomic-embed-text"
    Dim     int
}

func (p *OllamaEmbeddingProvider) CreateEmbedding(text string) ([]float32, error) {
    body := map[string]interface{}{"model": p.Model, "prompt": text}
    jsonBody, _ := json.Marshal(body)
    resp, err := http.Post(p.BaseURL+"/api/embeddings", "application/json", bytes.NewReader(jsonBody))
    // ... parse response
}
```

### 11. No OpenAPI/Swagger docs

**Where:** `api/routes/`

The API has no documentation. A consumer has to read the source code to understand request/response formats.

**Fix:** Use `github.com/swaggo/swag` with `net/http` support:

```go
// @Summary      Ingest a document
// @Description  Upload a document for chunking, embedding, and indexing
// @Tags         documents
// @Accept       json
// @Produce      json
// @Param        request body dto.IngestDocumentRequest true "Document content"
// @Success      201 {object} documents.IngestResult
// @Failure      400 {object} map[string]string
// @Router       /api/v1/documents [post]
func (ctrl *KnowledgeController) IngestDocument(w http.ResponseWriter, r *http.Request) {
```

Then serve the swagger UI at `/api/v1/docs`.

### 12. Version is hardcoded

**Where:** [`config/config.go`](config/config.go:72)

`Version: "1.0.0"` is hardcoded. It won't update with releases.

**Fix:** Inject at build time:

```go
// config/config.go
var Version = "dev" // overridden by -ldflags

// In Load():
Version: Version,
```

```makefile
# Makefile
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
build:
	go build -ldflags="-X github.com/akhmadzaki/go-backend/config.Version=$(VERSION) -s -w" -o go-backend .
```

---

## 🟢 Medium (Polish & Professionalism)

### 13. No Makefile

**Where:** Project root

No standardized build/run/test commands. Developers must remember Go commands manually.

**Fix:** Add a `Makefile`:

```makefile
.PHONY: build run test clean docker-build docker-up lint

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

build:
	go build -ldflags="-X github.com/akhmadzaki/go-backend/config.Version=$(VERSION) -s -w" -o go-backend .

run:
	go run main.go

test:
	go test ./... -v -race -count=1

clean:
	rm -f go-backend
	rm -rf build/

docker-build:
	docker compose build

docker-up:
	docker compose up -d

lint:
	go vet ./...
```

### 14. Health check is superficial

**Where:** [`api/controllers/knowledge_controller.go`](api/controllers/knowledge_controller.go:277)

The health endpoint only returns a static JSON. It doesn't verify that dependencies (DB, Ollama, vector store) are actually healthy.

**Fix:** Add a readiness probe that checks all dependencies:

```go
func (ctrl *KnowledgeController) HealthCheck(w http.ResponseWriter, r *http.Request) {
    checks := map[string]string{
        "database":    "ok",
        "ollama":      "ok",
        "vectorstore": "ok",
    }

    // Check DB
    if err := ctrl.db.PingContext(r.Context()); err != nil {
        checks["database"] = "unreachable"
    }

    // Check Ollama
    resp, err := http.Get(cfg.LLM.BaseURL + "/api/tags")
    if err != nil || resp.StatusCode != 200 {
        checks["ollama"] = "unreachable"
    }

    // Check vector store
    if !ctrl.vecStore.HealthCheck() {
        checks["vectorstore"] = "unhealthy"
    }

    status := http.StatusOK
    for _, v := range checks {
        if v != "ok" {
            status = http.StatusServiceUnavailable
            break
        }
    }

    helpers.RespondJSON(w, status, map[string]interface{}{
        "service": "local-ai-knowledge-base",
        "status":  checks,
        "version": cfg.App.Version,
    })
}
```

### 15. No request ID tracing

**Where:** `main.go` / middleware

Cannot correlate log entries across a single request flow.

**Fix:** Add a middleware that injects a request ID:

```go
func requestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := r.Header.Get("X-Request-ID")
        if id == "" {
            id = uuid.New().String()
        }
        w.Header().Set("X-Request-ID", id)
        ctx := context.WithValue(r.Context(), "request_id", id)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 16. Database pool config is hardcoded

**Where:** [`db/db.go`](db/db.go:32-33)

`SetMaxOpenConns(25)` and `SetMaxIdleConns(5)` are hardcoded.

**Fix:** Add to `DatabaseConfig`:

```go
type DatabaseConfig struct {
    Host         string
    Port         int
    Name         string
    User         string
    Password     string
    MaxOpenConns int
    MaxIdleConns int
    ConnMaxLifetime time.Duration
}
```

With env var defaults:

```go
MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25),
MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 5),
```

### 17. No Docker resource limits

**Where:** [`docker-compose.yml`](docker-compose.yml)

Containers can consume all host CPU/memory, causing OOM kills or resource starvation.

**Fix:** Add resource limits:

```yaml
services:
  postgres:
    ...
    deploy:
      resources:
        limits:
          cpus: "2"
          memory: "1G"

  backend:
    ...
    deploy:
      resources:
        limits:
          cpus: "1"
          memory: "512M"

  frontend:
    ...
    deploy:
      resources:
        limits:
          cpus: "0.5"
          memory: "256M"
```

### 18. Frontend lacks error boundaries & loading states

**Where:** `frontend/src/`

The React app has no error boundaries, no Suspense fallbacks, and minimal loading states.

**Fix:**

```tsx
// ErrorBoundary.tsx
class ErrorBoundary extends React.Component<
  { children: React.ReactNode; fallback?: React.ReactNode },
  { hasError: boolean }
> {
  state = { hasError: false };

  static getDerivedStateFromError() {
    return { hasError: true };
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback ?? <div>Something went wrong</div>;
    }
    return this.props.children;
  }
}
```

Add skeleton loaders for document list, search results, and chat messages.

### 19. Keyword search is a stub

**Where:** [`knowledge/search/search_service.go`](knowledge/search/search_service.go:90)

`keywordSearch` returns an empty response. The `chunks` table has a `tsv_content TSVECTOR` column (from migration 002), but it's never queried.

**Fix:** Implement PostgreSQL full-text search:

```go
func (svc *SearchService) keywordSearch(req SearchRequest) SearchResponse {
    query := `
        SELECT c.id, c.document_id, d.title, c.content,
               ts_rank(c.tsv_content, plainto_tsquery('english', $1)) AS score
        FROM chunks c
        JOIN documents d ON d.id = c.document_id
        WHERE c.tsv_content @@ plainto_tsquery('english', $1)
        ORDER BY score DESC
        LIMIT $2
    `
    rows, err := svc.vectorStore.DB().Query(query, req.Query, req.TopK)
    // ... scan results
}
```

### 20. API versioning

**Where:** `api/routes/knowledge_routes.go`

The `/api/v1/` prefix is already correct ✅. No action needed, but consider adding `Accept-Version` header support for future API evolution.

---

## 🟣 Low (Nice-to-Have Portfolio Highlights)

### 21. No metrics / observability

**Where:** `main.go`

No Prometheus metrics, no Grafana dashboards. You can't monitor request rates, error rates, or latency.

**Fix:** Add `github.com/prometheus/client_golang/prometheus/promhttp`:

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

// In RegisterRoutes:
mux.Handle("GET /metrics", promhttp.Handler())

// Wrap handlers with:
promMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        timer := prometheus.NewTimer(httpRequestDuration.WithLabelValues(r.URL.Path))
        defer timer.ObserveDuration()
        httpRequestsTotal.WithLabelValues(r.URL.Path, r.Method).Inc()
        next(w, r)
    }
}
```

### 22. No structured JSON logging in production

**Where:** [`logger/logger.go`](logger/logger.go:54)

The logger outputs human-readable text. In production, log aggregators need JSON.

**Fix:** In production mode, output JSON:

```go
func logEntry(severity, component, message string, args ...interface{}) {
    entry := map[string]interface{}{
        "timestamp":   time.Now().Format(time.RFC3339),
        "severity":    severity,
        "component":   component,
        "message":     message,
        "environment": os.Getenv("APP_ENV"),
    }
    if len(args) > 0 {
        entry["args"] = args
    }

    var output []byte
    if os.Getenv("APP_ENV") == "production" {
        output, _ = json.Marshal(entry)
    } else {
        output = []byte(fmt.Sprintf("[%s] [%s] [%s] %s", entry["timestamp"], severity, component, message))
    }

    if severity == "ERROR" {
        os.Stderr.Write(append(output, '\n'))
    } else {
        os.Stdout.Write(append(output, '\n'))
    }
}
```

### 23. No database migration tool

**Where:** `db/migrations/`

Migrations run on startup via custom code. There's no way to roll back, check migration status, or manage versions cleanly.

**Fix:** Integrate `github.com/golang-migrate/migrate/v4`:

```go
import (
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dsn string) error {
    m, err := migrate.New("file://db/migrations/sql", dsn)
    if err != nil {
        return err
    }
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }
    return nil
}
```

Then convert the `.go` migration files to SQL files:

```
db/migrations/sql/
  001_create_documents.up.sql
  001_create_documents.down.sql
  002_create_chunks.up.sql
  002_create_chunks.down.sql
  ...
```

### 24. No SSL/TLS

**Where:** `frontend/nginx.conf`

Traffic is plain HTTP. For any real deployment, you need TLS.

**Fix:** Add Let's Encrypt with `certbot` or use a reverse proxy (Caddy, Traefik) that handles TLS automatically.

### 25. No PWA support

**Where:** `frontend/`

The frontend can't work offline and can't be installed as a PWA.

**Fix:** Add a service worker with `vite-plugin-pwa`:

```ts
// vite.config.ts
import { VitePWA } from 'vite-plugin-pwa';

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: 'autoUpdate',
      manifest: {
        name: 'AI Knowledge Base',
        short_name: 'KnowledgeBase',
        theme_color: '#ffffff',
      },
    }),
  ],
});
```

### 26. Frontend lacks TypeScript strict mode

**Where:** `frontend/tsconfig.json`

Without `strict: true`, TypeScript allows `null`/`undefined` access that causes runtime errors.

**Fix:**

```json
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true
  }
}
```

### 27. No pre-commit hooks

**Where:** Project root

No automated linting or formatting before commits.

**Fix:** Add `husky` + `lint-staged`:

```bash
cd frontend
npx husky init
```

```json
// frontend/package.json
{
  "lint-staged": {
    "*.{ts,tsx}": ["eslint --fix", "prettier --write"],
    "*.css": ["prettier --write"]
  }
}
```

For Go:

```bash
# .husky/pre-commit
cd backend && go vet ./...
```
