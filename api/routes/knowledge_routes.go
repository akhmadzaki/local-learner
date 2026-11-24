package routes

import (
	"database/sql"
	"net/http"

	"github.com/akhmadzaki/go-backend/api/controllers"
	"github.com/akhmadzaki/go-backend/config"
	_ "github.com/akhmadzaki/go-backend/docs" // swagger generated docs
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/time/rate"
)

// RegisterRoutes wires up all HTTP endpoints for the knowledge-base API.
// Uses standard net/http mux with middleware for CORS, rate limiting,
// request ID tracing, and panic recovery.
func RegisterRoutes(cfg *config.Config, db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	ctrl := controllers.NewKnowledgeController(cfg, db)

	// ── Document CRUD ─────────────────────────────────────────────────────
	mux.HandleFunc("POST /api/v1/documents", ctrl.IngestDocument)
	mux.HandleFunc("GET /api/v1/documents", ctrl.ListDocuments)
	mux.HandleFunc("GET /api/v1/documents/{id}", ctrl.GetDocument)
	mux.HandleFunc("DELETE /api/v1/documents/{id}", ctrl.DeleteDocument)

	// ── Search ────────────────────────────────────────────────────────────
	mux.HandleFunc("POST /api/v1/search", ctrl.Search)
	mux.HandleFunc("POST /api/v1/search/vector", ctrl.VectorSearch)
	mux.HandleFunc("POST /api/v1/search/keyword", ctrl.KeywordSearch)
	mux.HandleFunc("POST /api/v1/search/hybrid", ctrl.HybridSearch)

	// ── Agent / Chat ──────────────────────────────────────────────────────
	mux.HandleFunc("POST /api/v1/chat", ctrl.Chat)
	mux.HandleFunc("POST /api/v1/chat/{convID}", ctrl.ChatContinue)
	mux.HandleFunc("GET /api/v1/chat/{convID}", ctrl.GetConversation)
	mux.HandleFunc("GET /api/v1/conversations", ctrl.ListConversations)
	mux.HandleFunc("DELETE /api/v1/chat/{convID}", ctrl.DeleteConversation)

	// ── Index Management ──────────────────────────────────────────────────
	mux.HandleFunc("POST /api/v1/index/build", ctrl.BuildIndex)
	mux.HandleFunc("GET /api/v1/index/stats", ctrl.GetIndexStats)

	// ── Pipeline ──────────────────────────────────────────────────────────
	mux.HandleFunc("POST /api/v1/pipeline/ingest", ctrl.RunPipeline)
	mux.HandleFunc("GET /api/v1/pipeline/jobs/{jobID}", ctrl.GetJobStatus)

	// ── Health ────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /api/v1/health", ctrl.HealthCheck)

	// ── Swagger UI ────────────────────────────────────────────────────────
	// Swagger documentation: GET /swagger/index.html
	mux.Handle("GET /swagger/", httpSwagger.Handler(httpSwagger.URL("doc.json")))

	// ── Apply middleware (outermost first) ────────────────────────────────
	var handler http.Handler = mux

	// Recovery must be outermost to catch panics from all layers
	handler = recoveryMiddleware(handler)

	// Request ID for tracing
	handler = requestIDMiddleware(handler)

	// CORS for cross-origin frontend access
	handler = corsMiddleware(handler)

	// Rate limiting: 100 requests/second per IP with burst of 200
	rateLimiter := NewRateLimiter(rate.Limit(100), 200)
	handler = rateLimiter.Middleware(handler)

	return handler
}
