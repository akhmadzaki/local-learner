package documents

import (
	"context"
	"testing"

	"github.com/akhmadzaki/go-backend/db/models"
	"github.com/akhmadzaki/go-backend/knowledge/embeddings"
)

// mockDocRepo implements DocumentRepository interface.
type mockDocRepo struct {
	createFn   func(ctx context.Context, title, source, status string) *models.Document
	findByIDFn func(ctx context.Context, id string) *models.Document
	findAllFn  func(ctx context.Context, offset, limit int) []*models.Document
	updateFn   func(ctx context.Context, id, status string) bool
	deleteFn   func(ctx context.Context, id string) bool
}

func (m *mockDocRepo) Create(ctx context.Context, title, source, status string) *models.Document {
	return m.createFn(ctx, title, source, status)
}

func (m *mockDocRepo) FindByID(ctx context.Context, id string) *models.Document {
	return m.findByIDFn(ctx, id)
}

func (m *mockDocRepo) FindAll(ctx context.Context, offset, limit int) []*models.Document {
	return m.findAllFn(ctx, offset, limit)
}

func (m *mockDocRepo) UpdateStatus(ctx context.Context, id, status string) bool {
	return m.updateFn(ctx, id, status)
}

func (m *mockDocRepo) Delete(ctx context.Context, id string) bool {
	return m.deleteFn(ctx, id)
}

// mockChunkSvc implements ChunkServiceInterface.
type mockChunkSvc struct {
	splitFn  func(ctx context.Context, content, docID string) []*models.Chunk
	deleteFn func(ctx context.Context, docID string) bool
}

func (m *mockChunkSvc) SplitContent(ctx context.Context, content, docID string) []*models.Chunk {
	return m.splitFn(ctx, content, docID)
}

func (m *mockChunkSvc) DeleteChunksByDocument(ctx context.Context, docID string) bool {
	return m.deleteFn(ctx, docID)
}

// mockEmbedSvc implements EmbeddingServiceInterface.
type mockEmbedSvc struct {
	generateFn func(ctx context.Context, chunks []*models.Chunk) []embeddings.EmbeddingResult
	storeFn    func(ctx context.Context, results []embeddings.EmbeddingResult)
	deleteFn   func(ctx context.Context, docID string)
}

func (m *mockEmbedSvc) GenerateEmbeddings(ctx context.Context, chunks []*models.Chunk) []embeddings.EmbeddingResult {
	return m.generateFn(ctx, chunks)
}

func (m *mockEmbedSvc) StoreVectors(ctx context.Context, results []embeddings.EmbeddingResult) {
	m.storeFn(ctx, results)
}

func (m *mockEmbedSvc) DeleteVectorsByDocument(ctx context.Context, docID string) {
	m.deleteFn(ctx, docID)
}

func TestIngestDocument_Success(t *testing.T) {
	ctx := context.Background()

	mockRepo := &mockDocRepo{
		createFn: func(ctx context.Context, title, source, status string) *models.Document {
			return &models.Document{ID: "doc-1", Title: title, Source: source, Status: status}
		},
		updateFn: func(ctx context.Context, id, status string) bool {
			return true
		},
	}

	mockChunk := &mockChunkSvc{
		splitFn: func(ctx context.Context, content, docID string) []*models.Chunk {
			return []*models.Chunk{
				{ID: "chunk-1", DocumentID: docID, Content: content[:10]},
			}
		},
	}

	mockEmbed := &mockEmbedSvc{
		generateFn: func(ctx context.Context, chunks []*models.Chunk) []embeddings.EmbeddingResult {
			return []embeddings.EmbeddingResult{
				{ChunkID: chunks[0].ID, Vector: []float32{0.1, 0.2, 0.3}, ModelName: "test-model"},
			}
		},
		storeFn: func(ctx context.Context, results []embeddings.EmbeddingResult) {},
	}

	svc := &DocumentService{
		docRepo:  mockRepo,
		chunkSvc: mockChunk,
		embedSvc: mockEmbed,
	}

	result := svc.IngestDocument(ctx, "Test Title", "Hello World Content", "test")

	if result.Status != "indexed" {
		t.Errorf("expected status 'indexed', got '%s'", result.Status)
	}
	if result.DocumentID != "doc-1" {
		t.Errorf("expected document ID 'doc-1', got '%s'", result.DocumentID)
	}
	if result.ChunkCount != 1 {
		t.Errorf("expected chunk count 1, got %d", result.ChunkCount)
	}
}

func TestIngestDocument_CreateFails(t *testing.T) {
	ctx := context.Background()

	mockRepo := &mockDocRepo{
		createFn: func(ctx context.Context, title, source, status string) *models.Document {
			return nil
		},
	}

	svc := &DocumentService{
		docRepo: mockRepo,
	}

	result := svc.IngestDocument(ctx, "Test", "Content", "test")

	if result.Status != "failed" {
		t.Errorf("expected status 'failed', got '%s'", result.Status)
	}
	if result.Error == "" {
		t.Error("expected error message, got empty")
	}
}

func TestDeleteDocument_Success(t *testing.T) {
	ctx := context.Background()

	deleted := false
	mockRepo := &mockDocRepo{
		deleteFn: func(ctx context.Context, id string) bool {
			deleted = true
			return true
		},
	}

	mockChunk := &mockChunkSvc{
		deleteFn: func(ctx context.Context, docID string) bool {
			return true
		},
	}

	mockEmbed := &mockEmbedSvc{
		deleteFn: func(ctx context.Context, docID string) {},
	}

	svc := &DocumentService{
		docRepo:  mockRepo,
		chunkSvc: mockChunk,
		embedSvc: mockEmbed,
	}

	result := svc.DeleteDocument(ctx, "doc-1")
	if !result {
		t.Error("expected DeleteDocument to return true")
	}
	if !deleted {
		t.Error("expected repository Delete to be called")
	}
}

func TestGetDocument(t *testing.T) {
	ctx := context.Background()

	expectedDoc := &models.Document{ID: "doc-1", Title: "Test Doc"}
	mockRepo := &mockDocRepo{
		findByIDFn: func(ctx context.Context, id string) *models.Document {
			if id == "doc-1" {
				return expectedDoc
			}
			return nil
		},
	}

	svc := &DocumentService{docRepo: mockRepo}

	doc := svc.GetDocument(ctx, "doc-1")
	if doc == nil {
		t.Fatal("expected document, got nil")
	}
	if doc.ID != "doc-1" {
		t.Errorf("expected doc-1, got %s", doc.ID)
	}

	doc = svc.GetDocument(ctx, "nonexistent")
	if doc != nil {
		t.Error("expected nil for nonexistent document")
	}
}

func TestListDocuments(t *testing.T) {
	ctx := context.Background()

	expectedDocs := []*models.Document{
		{ID: "doc-1", Title: "Doc 1"},
		{ID: "doc-2", Title: "Doc 2"},
	}
	mockRepo := &mockDocRepo{
		findAllFn: func(ctx context.Context, offset, limit int) []*models.Document {
			if offset == 0 && limit == 20 {
				return expectedDocs
			}
			return nil
		},
	}

	svc := &DocumentService{docRepo: mockRepo}

	docs := svc.ListDocuments(ctx, 0, 20)
	if len(docs) != 2 {
		t.Errorf("expected 2 documents, got %d", len(docs))
	}
}
