package controllers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/akhmadzaki/go-backend/config"
	"github.com/akhmadzaki/go-backend/db/repository"
	"github.com/akhmadzaki/go-backend/dto"
	"github.com/akhmadzaki/go-backend/helpers"
	"github.com/akhmadzaki/go-backend/knowledge/agent"
	"github.com/akhmadzaki/go-backend/knowledge/chunks"
	"github.com/akhmadzaki/go-backend/knowledge/documents"
	"github.com/akhmadzaki/go-backend/knowledge/embeddings"
	"github.com/akhmadzaki/go-backend/knowledge/index"
	"github.com/akhmadzaki/go-backend/knowledge/pipeline"
	"github.com/akhmadzaki/go-backend/knowledge/search"
	"github.com/akhmadzaki/go-backend/knowledge/vectorstore"
)

// KnowledgeController handles HTTP requests for the knowledge base.
type KnowledgeController struct {
	cfg       *config.Config
	db        *sql.DB
	docSvc    *documents.DocumentService
	searchSvc *search.SearchService
	agentSvc  *agent.AgentService
	idxSvc    *index.IndexService
	vecStore  *vectorstore.VectorStore
	pipeline  *pipeline.RAGPipeline
	orch      *pipeline.PipelineOrchestrator
	docProc   *documents.DocumentProcessor
}

// NewKnowledgeController creates a fully wired controller with all dependencies
// constructed from the provided config and database connection.
func NewKnowledgeController(cfg *config.Config, db *sql.DB) *KnowledgeController {
	// ── Repositories ──────────────────────────────────────────────────────
	docRepo := repository.NewDocumentRepository(db)
	chunkRepo := repository.NewChunkRepository(db)
	embedRepo := repository.NewEmbeddingRepository(db)
	convRepo := repository.NewConversationRepository(db)

	// ── Embedding provider ────────────────────────────────────────────────
	// API key is read from environment; pass empty string to let the provider
	// factory fall back to env-based lookup if needed.
	embedProvider := embeddings.NewEmbeddingProvider(
		cfg.Embeddings.Provider,
		cfg.Embeddings.Model,
		cfg.Embeddings.Dim,
		cfg.Embeddings.BaseURL,
		"",
	)

	// ── Vector store ──────────────────────────────────────────────────────
	vecStore := vectorstore.NewVectorStore(db, "embeddings", cfg.Embeddings.Dim)

	// ── Embedding service ─────────────────────────────────────────────────
	embedSvc := embeddings.NewEmbeddingService(embedProvider, vecStore, embedRepo)

	// ── Chunk strategy ────────────────────────────────────────────────────
	var strategy chunks.ChunkStrategy
	switch cfg.Chunking.Strategy {
	case "fixed":
		strategy = chunks.NewFixedSizeChunker(cfg.Chunking.ChunkSize, cfg.Chunking.ChunkOverlap)
	case "sentence":
		strategy = chunks.NewSentenceChunker(cfg.Chunking.ChunkSize, cfg.Chunking.ChunkOverlap)
	case "paragraph":
		strategy = chunks.NewParagraphChunker()
	default:
		log.Printf("KnowledgeController: unknown chunk strategy '%s', falling back to 'sentence'", cfg.Chunking.Strategy)
		strategy = chunks.NewSentenceChunker(cfg.Chunking.ChunkSize, cfg.Chunking.ChunkOverlap)
	}

	chunkConfig := chunks.ChunkConfig{
		Strategy:       cfg.Chunking.Strategy,
		ChunkSize:      cfg.Chunking.ChunkSize,
		ChunkOverlap:   cfg.Chunking.ChunkOverlap,
		MinChunkLength: cfg.Chunking.MinChunkLength,
	}

	// ── Chunk service ─────────────────────────────────────────────────────
	chunkSvc := chunks.NewChunkService(chunkRepo, strategy, chunkConfig)

	// ── Document service ──────────────────────────────────────────────────
	docSvc := documents.NewDocumentService(docRepo, chunkSvc, embedSvc)

	// ── Search service ────────────────────────────────────────────────────
	searchSvc := search.NewSearchService(embedSvc, vecStore, db)

	// ── Index service ─────────────────────────────────────────────────────
	idxSvc := index.NewIndexService(embedSvc, vecStore, db)

	// ── Agent / Chat ──────────────────────────────────────────────────────
	convManager := agent.NewConversationManager(convRepo, cfg.LLM.MaxHistory)
	agentSvc := agent.NewAgentService(
		searchSvc,
		convManager,
		cfg.LLM.Model,
		cfg.LLM.MaxTokens,
		cfg.LLM.Temperature,
		cfg.LLM.BaseURL,
	)

	// ── Pipeline ──────────────────────────────────────────────────────────
	ragPipeline := pipeline.NewRAGPipeline(docSvc, chunkSvc, embedSvc, idxSvc)
	orch := pipeline.NewPipelineOrchestrator(ragPipeline, 3)

	// ── Document processor ──────────────────────────────────────────────────
	docProc := documents.NewDocumentProcessor(cfg.Upload.MaxFileSize)

	log.Println("KnowledgeController: all dependencies wired successfully")
	return &KnowledgeController{
		cfg:       cfg,
		db:        db,
		docSvc:    docSvc,
		searchSvc: searchSvc,
		agentSvc:  agentSvc,
		idxSvc:    idxSvc,
		vecStore:  vecStore,
		pipeline:  ragPipeline,
		orch:      orch,
		docProc:   docProc,
	}
}

// ── Document endpoints ───────────────────────────────────────────────────────

// IngestDocument godoc
// @Summary      Ingest a new document
// @Description  Accepts raw text content, splits it into chunks, generates
// @Description  embeddings, and stores everything for hybrid search.
// @Tags         documents
// @Accept       json
// @Produce      json
// @Param        request  body      dto.IngestDocumentRequest  true  "Document content to ingest"
// @Success      201      {object}  documents.IngestResult     "Ingested document summary"
// @Failure      400      {object}  helpers.ErrorResponse      "Invalid request body"
// @Failure      500      {object}  helpers.ErrorResponse      "Internal server error"
// @Router       /documents [post]
func (ctrl *KnowledgeController) IngestDocument(w http.ResponseWriter, r *http.Request) {
	var req dto.IngestDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result := ctrl.docSvc.IngestDocument(r.Context(), req.Title, req.Content, req.Source)
	helpers.RespondCreated(w, result)
}

// UploadDocument godoc
// @Summary      Upload a document file (PDF, Markdown, HTML, TXT)
// @Description  Accepts a file upload via multipart/form-data, extracts text
// @Description  content from the file, then ingests it into the knowledge base.
// @Tags         documents
// @Accept       mpfd
// @Produce      json
// @Param        file    formData  file    true  "Document file to upload (pdf, md, html, txt)"
// @Param        title   formData  string false "Optional document title (defaults to filename)"
// @Param        source  formData  string false "Optional source label (defaults to 'upload')"
// @Success      201     {object}  documents.IngestResult  "Ingested document summary"
// @Failure      400     {object}  helpers.ErrorResponse   "Invalid file or missing data"
// @Failure      413     {object}  helpers.ErrorResponse   "File too large"
// @Failure      500     {object}  helpers.ErrorResponse   "Internal server error"
// @Router       /documents/upload [post]
func (ctrl *KnowledgeController) UploadDocument(w http.ResponseWriter, r *http.Request) {
	// Limit request body size to MaxFileSize + a small buffer for form overhead
	r.Body = http.MaxBytesReader(w, r.Body, int64(ctrl.cfg.Upload.MaxFileSize)+1024*1024)

	if err := r.ParseMultipartForm(int64(ctrl.cfg.Upload.MaxFileSize) + 1024*1024); err != nil {
		helpers.RespondError(w, http.StatusBadRequest, "File too large or invalid multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		helpers.RespondError(w, http.StatusBadRequest, "Missing file in request")
		return
	}
	defer file.Close()

	// Read the uploaded file into memory
	data := make([]byte, header.Size)
	if _, err := file.Read(data); err != nil {
		helpers.RespondError(w, http.StatusInternalServerError, "Failed to read uploaded file")
		return
	}

	// Determine title: use explicit form field or fall back to filename
	title := r.FormValue("title")
	if title == "" {
		title = header.Filename
	}

	source := r.FormValue("source")
	if source == "" {
		source = "upload"
	}

	// Process the file through the document processor (extract text, validate)
	result := ctrl.docProc.ProcessFile(header.Filename, data)
	if result == nil {
		helpers.RespondError(w, http.StatusBadRequest, "Unsupported file type or file exceeds size limit")
		return
	}

	if result.Content == "" {
		helpers.RespondError(w, http.StatusBadRequest, "No text content could be extracted from the file")
		return
	}

	// Ingest the extracted content into the knowledge base
	ingestResult := ctrl.docSvc.IngestDocument(r.Context(), title, result.Content, source)
	helpers.RespondCreated(w, ingestResult)
}

// ListDocuments godoc
// @Summary      List documents
// @Description  Returns a paginated list of documents in the knowledge base.
// @Tags         documents
// @Produce      json
// @Success      200  {array}   models.Document          "List of documents"
// @Failure      500  {object}  helpers.ErrorResponse    "Internal server error"
// @Router       /documents [get]
func (ctrl *KnowledgeController) ListDocuments(w http.ResponseWriter, r *http.Request) {
	docs := ctrl.docSvc.ListDocuments(r.Context(), 0, 20)
	helpers.RespondSuccess(w, docs)
}

// GetDocument godoc
// @Summary      Get a document by ID
// @Description  Returns a single document and its metadata.
// @Tags         documents
// @Produce      json
// @Param        id    path      string                true  "Document UUID"
// @Success      200   {object}  models.Document       "Document details"
// @Failure      404   {object}  helpers.ErrorResponse "Document not found"
// @Router       /documents/{id} [get]
func (ctrl *KnowledgeController) GetDocument(w http.ResponseWriter, r *http.Request) {
	docID := r.PathValue("id")
	doc := ctrl.docSvc.GetDocument(r.Context(), docID)
	if doc == nil {
		helpers.RespondError(w, http.StatusNotFound, "Document not found")
		return
	}
	helpers.RespondSuccess(w, doc)
}

// DeleteDocument godoc
// @Summary      Delete a document
// @Description  Removes a document and all its associated chunks and vectors.
// @Tags         documents
// @Produce      json
// @Param        id    path      string                true  "Document UUID"
// @Success      200   {object}  map[string]bool       "Deletion confirmation"
// @Failure      404   {object}  helpers.ErrorResponse "Document not found"
// @Router       /documents/{id} [delete]
func (ctrl *KnowledgeController) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	docID := r.PathValue("id")
	if ctrl.docSvc.DeleteDocument(r.Context(), docID) {
		helpers.RespondJSON(w, http.StatusOK, map[string]bool{"deleted": true})
	} else {
		helpers.RespondError(w, http.StatusNotFound, "Document not found")
	}
}

// ── Search endpoints ─────────────────────────────────────────────────────────

// Search godoc
// @Summary      Search the knowledge base
// @Description  Performs a search over the knowledge base using the mode
// @Description  specified in the request (vector, keyword, or hybrid).
// @Tags         search
// @Accept       json
// @Produce      json
// @Param        request  body      search.SearchRequest     true  "Search parameters"
// @Success      200      {object}  search.SearchResponse    "Ranked search results"
// @Failure      400      {object}  helpers.ErrorResponse    "Invalid request body"
// @Failure      500      {object}  helpers.ErrorResponse    "Internal server error"
// @Router       /search [post]
func (ctrl *KnowledgeController) Search(w http.ResponseWriter, r *http.Request) {
	var req search.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp := ctrl.searchSvc.Search(r.Context(), req)
	helpers.RespondSuccess(w, resp)
}

// VectorSearch godoc
// @Summary      Vector search
// @Description  Performs pure semantic (embedding-based) vector search.
// @Tags         search
// @Accept       json
// @Produce      json
// @Param        request  body      search.SearchRequest     true  "Search parameters (mode is overridden to vector)"
// @Success      200      {object}  search.SearchResponse    "Vector search results"
// @Failure      400      {object}  helpers.ErrorResponse    "Invalid request body"
// @Router       /search/vector [post]
func (ctrl *KnowledgeController) VectorSearch(w http.ResponseWriter, r *http.Request) {
	var req search.SearchRequest
	json.NewDecoder(r.Body).Decode(&req)
	req.Mode = "vector"
	resp := ctrl.searchSvc.Search(r.Context(), req)
	helpers.RespondSuccess(w, resp)
}

// KeywordSearch godoc
// @Summary      Keyword search
// @Description  Performs full-text / BM25 style keyword search via PostgreSQL.
// @Tags         search
// @Accept       json
// @Produce      json
// @Param        request  body      search.SearchRequest     true  "Search parameters (mode is overridden to keyword)"
// @Success      200      {object}  search.SearchResponse    "Keyword search results"
// @Failure      400      {object}  helpers.ErrorResponse    "Invalid request body"
// @Router       /search/keyword [post]
func (ctrl *KnowledgeController) KeywordSearch(w http.ResponseWriter, r *http.Request) {
	var req search.SearchRequest
	json.NewDecoder(r.Body).Decode(&req)
	req.Mode = "keyword"
	resp := ctrl.searchSvc.Search(r.Context(), req)
	helpers.RespondSuccess(w, resp)
}

// HybridSearch godoc
// @Summary      Hybrid search
// @Description  Combines vector and keyword retrieval with configurable weighting.
// @Tags         search
// @Accept       json
// @Produce      json
// @Param        request  body      search.SearchRequest     true  "Search parameters (mode is overridden to hybrid)"
// @Success      200      {object}  search.SearchResponse    "Hybrid search results"
// @Failure      400      {object}  helpers.ErrorResponse    "Invalid request body"
// @Router       /search/hybrid [post]
func (ctrl *KnowledgeController) HybridSearch(w http.ResponseWriter, r *http.Request) {
	var req search.SearchRequest
	json.NewDecoder(r.Body).Decode(&req)
	req.Mode = "hybrid"
	resp := ctrl.searchSvc.Search(r.Context(), req)
	helpers.RespondSuccess(w, resp)
}

// ── Chat / Agent endpoints ───────────────────────────────────────────────────

// Chat godoc
// @Summary      Start a new chat
// @Description  Sends a message to the agent, optionally continuing an existing
// @Description  conversation if a conversation_id is provided.
// @Tags         chat
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ChatRequestDTO       true  "Chat message"
// @Success      200      {object}  agent.AgentResponse      "Agent reply with sources and reasoning"
// @Failure      400      {object}  helpers.ErrorResponse    "Invalid request body"
// @Failure      500      {object}  helpers.ErrorResponse    "Internal server error"
// @Router       /chat [post]
func (ctrl *KnowledgeController) Chat(w http.ResponseWriter, r *http.Request) {
	var req dto.ChatRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	agentReq := agent.AgentRequest{
		Query:          req.Message,
		ConversationID: req.ConversationID,
		Mode:           req.Mode,
	}
	resp := ctrl.agentSvc.ProcessMessage(r.Context(), agentReq)
	helpers.RespondSuccess(w, resp)
}

// ChatContinue godoc
// @Summary      Continue a conversation
// @Description  Sends a follow-up message within an existing conversation.
// @Tags         chat
// @Accept       json
// @Produce      json
// @Param        convID   path      string                true  "Conversation UUID"
// @Param        request  body      dto.ChatRequestDTO    true  "Chat message"
// @Success      200      {object}  agent.AgentResponse   "Agent reply with sources and reasoning"
// @Failure      400      {object}  helpers.ErrorResponse "Invalid request body"
// @Failure      404      {object}  helpers.ErrorResponse "Conversation not found"
// @Router       /chat/{convID} [post]
func (ctrl *KnowledgeController) ChatContinue(w http.ResponseWriter, r *http.Request) {
	convID := r.PathValue("convID")
	var req dto.ChatRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	agentReq := agent.AgentRequest{
		Query:          req.Message,
		ConversationID: convID,
		Mode:           req.Mode,
	}
	resp := ctrl.agentSvc.ProcessMessage(r.Context(), agentReq)
	helpers.RespondSuccess(w, resp)
}

// GetConversation godoc
// @Summary      Get conversation history
// @Description  Returns the full message history for a conversation.
// @Tags         chat
// @Produce      json
// @Param        convID   path      string                true  "Conversation UUID"
// @Success      200      {array}   agent.ConversationTurn  "Conversation history"
// @Failure      404      {object}  helpers.ErrorResponse   "Conversation not found"
// @Router       /chat/{convID} [get]
func (ctrl *KnowledgeController) GetConversation(w http.ResponseWriter, r *http.Request) {
	convID := r.PathValue("convID")
	history := ctrl.agentSvc.GetHistory(r.Context(), convID)
	helpers.RespondSuccess(w, history)
}

// ListConversations godoc
// @Summary      List conversations
// @Description  Returns all conversations with metadata.
// @Tags         chat
// @Produce      json
// @Success      200  {array}   map[string]interface{}    "List of conversations"
// @Failure      500  {object}  helpers.ErrorResponse      "Internal server error"
// @Router       /conversations [get]
func (ctrl *KnowledgeController) ListConversations(w http.ResponseWriter, r *http.Request) {
	convs := ctrl.agentSvc.ListConversations(r.Context())
	helpers.RespondSuccess(w, convs)
}

// DeleteConversation godoc
// @Summary      Delete a conversation
// @Description  Removes a conversation and all its stored turns.
// @Tags         chat
// @Produce      json
// @Param        convID   path      string                true  "Conversation UUID"
// @Success      200      {object}  map[string]bool       "Deletion confirmation"
// @Failure      404      {object}  helpers.ErrorResponse "Conversation not found"
// @Router       /chat/{convID} [delete]
func (ctrl *KnowledgeController) DeleteConversation(w http.ResponseWriter, r *http.Request) {
	convID := r.PathValue("convID")
	ctrl.agentSvc.DeleteConversation(r.Context(), convID)
	helpers.RespondJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

// ── Index endpoints ──────────────────────────────────────────────────────────

// BuildIndex godoc
// @Summary      Build search index
// @Description  Triggers a full re-index of all documents in the vector store.
// @Tags         index
// @Produce      json
// @Success      200  {object}  map[string]bool       "Indexing confirmation"
// @Failure      500  {object}  helpers.ErrorResponse "Internal server error"
// @Router       /index/build [post]
func (ctrl *KnowledgeController) BuildIndex(w http.ResponseWriter, r *http.Request) {
	success := ctrl.idxSvc.BuildIndex()
	helpers.RespondJSON(w, http.StatusOK, map[string]bool{"indexing": success})
}

// GetIndexStats godoc
// @Summary      Get index statistics
// @Description  Returns counts and health of the current vector index.
// @Tags         index
// @Produce      json
// @Success      200  {object}  index.IndexStats       "Index statistics"
// @Failure      500  {object}  helpers.ErrorResponse  "Internal server error"
// @Router       /index/stats [get]
func (ctrl *KnowledgeController) GetIndexStats(w http.ResponseWriter, r *http.Request) {
	stats := ctrl.idxSvc.GetStats()
	helpers.RespondSuccess(w, stats)
}

// ── Pipeline endpoints ───────────────────────────────────────────────────────

// RunPipeline godoc
// @Summary      Queue a pipeline ingestion job
// @Description  Enqueues an async ingestion job that processes a document
// @Description  through the full pipeline (chunking, embeddings, indexing).
// @Tags         pipeline
// @Accept       json
// @Produce      json
// @Param        request  body      dto.IngestDocumentRequest  true  "Document to process"
// @Success      202      {object}  map[string]string          "Queued job details"
// @Failure      400      {object}  helpers.ErrorResponse      "Invalid request body"
// @Router       /pipeline/ingest [post]
func (ctrl *KnowledgeController) RunPipeline(w http.ResponseWriter, r *http.Request) {
	var req dto.IngestDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	jobID := ctrl.orch.EnqueueJob(req.Title, "ingest")
	helpers.RespondJSON(w, http.StatusAccepted, map[string]string{"job_id": jobID, "status": "queued"})
}

// GetJobStatus godoc
// @Summary      Get pipeline job status
// @Description  Returns the current status of an async pipeline job.
// @Tags         pipeline
// @Produce      json
// @Param        jobID   path      string                true  "Pipeline job UUID"
// @Success      200     {object}  map[string]string     "Job status"
// @Failure      404     {object}  helpers.ErrorResponse "Job not found"
// @Router       /pipeline/jobs/{jobID} [get]
func (ctrl *KnowledgeController) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("jobID")
	status := ctrl.orch.GetJobStatus(jobID)
	helpers.RespondSuccess(w, map[string]string{"job_id": jobID, "status": status})
}

// ── Health ───────────────────────────────────────────────────────────────────

// HealthCheck godoc
// @Summary      Health check
// @Description  Returns service health, dependency status, and version.
// @Tags         system
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Service health"
// @Failure      503  {object}  map[string]interface{}  "Service degraded"
// @Router       /health [get]
func (ctrl *KnowledgeController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	checks := map[string]string{
		"database":    "ok",
		"ollama":      "ok",
		"vectorstore": "ok",
	}

	// Check database
	if err := ctrl.db.PingContext(ctx); err != nil {
		checks["database"] = "unreachable"
	}

	// Check vector store (uses the same DB connection)
	if !ctrl.vecStore.HealthCheck(ctx) {
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
		"version": ctrl.cfg.App.Version,
	})
}
