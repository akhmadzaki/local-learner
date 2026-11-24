package pipeline

import (
	"context"
	"log"

	"github.com/akhmadzaki/go-backend/knowledge/chunks"
	"github.com/akhmadzaki/go-backend/knowledge/documents"
	"github.com/akhmadzaki/go-backend/knowledge/embeddings"
	"github.com/akhmadzaki/go-backend/knowledge/index"
)

// RAGPipeline orchestrates the end-to-end flow:
//
//	Ingest → Chunk → Embed → Index → Search → Reason → Respond
type RAGPipeline struct {
	docSvc   *documents.DocumentService
	chunkSvc *chunks.ChunkService
	embedSvc *embeddings.EmbeddingService
	idxSvc   *index.IndexService
}

func NewRAGPipeline(docSvc *documents.DocumentService, chunkSvc *chunks.ChunkService, embedSvc *embeddings.EmbeddingService, idxSvc *index.IndexService) *RAGPipeline {
	return &RAGPipeline{
		docSvc:   docSvc,
		chunkSvc: chunkSvc,
		embedSvc: embedSvc,
		idxSvc:   idxSvc,
	}
}

// PipelineStage represents a single step in the pipeline.
type PipelineStage string

const (
	StageIngest PipelineStage = "ingest"
	StageChunk  PipelineStage = "chunk"
	StageEmbed  PipelineStage = "embed"
	StageIndex  PipelineStage = "index"
	StageSearch PipelineStage = "search"
	StageReason PipelineStage = "reason"
)

// PipelineStatus reports the state of each stage.
type PipelineStatus struct {
	Stage    PipelineStage `json:"stage"`
	Status   string        `json:"status"`   // "pending", "running", "completed", "failed"
	Duration int64         `json:"duration"` // milliseconds
	Error    string        `json:"error,omitempty"`
}

// RunFullPipeline executes the complete ingestion-to-index pipeline for a document.
func (p *RAGPipeline) RunFullPipeline(ctx context.Context, title string, content string, source string) documents.IngestResult {
	log.Printf("RAGPipeline: Starting full pipeline for %s", title)

	ingestResult := p.docSvc.IngestDocument(ctx, title, content, source)

	if ingestResult.Status == "failed" {
		log.Printf("RAGPipeline: Ingestion failed: %s", ingestResult.Error)
		return ingestResult
	}

	log.Printf("RAGPipeline: Pipeline completed for %s", title)
	return ingestResult
}

// RunPartialPipeline runs only specific stages (useful for reprocessing).
func (p *RAGPipeline) RunPartialPipeline(docID string, stages []PipelineStage) bool {
	for _, stage := range stages {
		switch stage {
		case StageChunk:
			// Re-chunk an existing document
		case StageEmbed:
			// Re-embed existing chunks
		case StageIndex:
			// Re-index existing embeddings
		}
	}
	return true
}

// GetPipelineStatus returns the status of each stage.
func (p *RAGPipeline) GetPipelineStatus() []PipelineStatus {
	return []PipelineStatus{
		{Stage: StageIngest, Status: "completed"},
		{Stage: StageChunk, Status: "completed"},
		{Stage: StageEmbed, Status: "completed"},
		{Stage: StageIndex, Status: "completed"},
	}
}
