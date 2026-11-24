package documents

import (
	"context"
	"log"

	"github.com/akhmadzaki/go-backend/db/models"
	"github.com/akhmadzaki/go-backend/db/repository"
	"github.com/akhmadzaki/go-backend/knowledge/chunks"
	"github.com/akhmadzaki/go-backend/knowledge/embeddings"
)

// DocumentRepository defines the interface for document persistence.
type DocumentRepository interface {
	Create(ctx context.Context, title string, source string, status string) *models.Document
	FindByID(ctx context.Context, id string) *models.Document
	FindAll(ctx context.Context, offset int, limit int) []*models.Document
	UpdateStatus(ctx context.Context, id string, status string) bool
	UpdateChunkCount(ctx context.Context, id string, count int) bool
	Delete(ctx context.Context, id string) bool
}

// ChunkServiceInterface defines the interface for chunk operations.
type ChunkServiceInterface interface {
	SplitContent(ctx context.Context, content string, docID string) []*models.Chunk
	DeleteChunksByDocument(ctx context.Context, docID string) bool
}

// EmbeddingServiceInterface defines the interface for embedding operations.
type EmbeddingServiceInterface interface {
	GenerateEmbeddings(ctx context.Context, chunks []*models.Chunk) []embeddings.EmbeddingResult
	StoreVectors(ctx context.Context, results []embeddings.EmbeddingResult) bool
	DeleteVectorsByDocument(ctx context.Context, docID string)
}

// DocumentService handles ingestion, processing, and lifecycle of knowledge-base documents.
type DocumentService struct {
	docRepo  DocumentRepository
	chunkSvc ChunkServiceInterface
	embedSvc EmbeddingServiceInterface
}

func NewDocumentService(docRepo *repository.DocumentRepository, chunkSvc *chunks.ChunkService, embedSvc *embeddings.EmbeddingService) *DocumentService {
	return &DocumentService{
		docRepo:  docRepo,
		chunkSvc: chunkSvc,
		embedSvc: embedSvc,
	}
}

// IngestResult summarises what happened during document ingestion.
type IngestResult struct {
	DocumentID string `json:"document_id"`
	ChunkCount int    `json:"chunk_count"`
	Status     string `json:"status"` // "indexed", "partial", "failed"
	Error      string `json:"error,omitempty"`
}

// IngestDocument accepts raw content, splits it into chunks, generates embeddings,
// and persists everything to both the relational DB and the vector store.
func (svc *DocumentService) IngestDocument(ctx context.Context, title string, content string, source string) IngestResult {
	doc := svc.docRepo.Create(ctx, title, source, "processing")
	if doc == nil {
		return IngestResult{Status: "failed", Error: "failed to create document"}
	}

	chunks := svc.chunkSvc.SplitContent(ctx, content, doc.ID)
	log.Printf("Document splitted")

	// Update chunk count on the document
	svc.docRepo.UpdateChunkCount(ctx, doc.ID, len(chunks))

	embedResults := svc.embedSvc.GenerateEmbeddings(ctx, chunks)
	log.Printf("Embeddings generated")

	if len(embedResults) == 0 {
		log.Printf("DocumentService: No embeddings generated for document %s", doc.ID)
		svc.docRepo.UpdateStatus(ctx, doc.ID, "partial")
		return IngestResult{DocumentID: doc.ID, ChunkCount: len(chunks), Status: "partial", Error: "no embeddings generated"}
	}

	if !svc.embedSvc.StoreVectors(ctx, embedResults) {
		log.Printf("DocumentService: Failed to store vectors for document %s", doc.ID)
		svc.docRepo.UpdateStatus(ctx, doc.ID, "partial")
		return IngestResult{DocumentID: doc.ID, ChunkCount: len(chunks), Status: "partial", Error: "failed to store vectors"}
	}
	log.Printf("Vector stored")

	svc.docRepo.UpdateStatus(ctx, doc.ID, "indexed")

	log.Printf("DocumentService: Ingested document %s (%d chunks)", doc.ID, len(chunks))
	return IngestResult{DocumentID: doc.ID, ChunkCount: len(chunks), Status: "indexed"}
}

// DeleteDocument removes a document and all its associated chunks and vectors.
func (svc *DocumentService) DeleteDocument(ctx context.Context, docID string) bool {
	svc.embedSvc.DeleteVectorsByDocument(ctx, docID)
	svc.chunkSvc.DeleteChunksByDocument(ctx, docID)
	return svc.docRepo.Delete(ctx, docID)
}

// GetDocument retrieves a document by its ID.
func (svc *DocumentService) GetDocument(ctx context.Context, docID string) *models.Document {
	return svc.docRepo.FindByID(ctx, docID)
}

// ListDocuments returns all documents with pagination.
func (svc *DocumentService) ListDocuments(ctx context.Context, offset int, limit int) []*models.Document {
	return svc.docRepo.FindAll(ctx, offset, limit)
}
