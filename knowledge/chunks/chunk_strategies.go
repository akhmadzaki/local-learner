package chunks

import (
	"regexp"
	"strings"

	"github.com/akhmadzaki/go-backend/db/models"
)

// ── Fixed-size chunking ──────────────────────────────────────────────────────
// Splits text into chunks of approximately `ChunkSize` characters with overlap.

type FixedSizeChunker struct {
	ChunkSize    int
	ChunkOverlap int
}

func NewFixedSizeChunker(chunkSize, chunkOverlap int) *FixedSizeChunker {
	return &FixedSizeChunker{
		ChunkSize:    chunkSize,
		ChunkOverlap: chunkOverlap,
	}
}

func (c *FixedSizeChunker) Split(text string, docID string, meta map[string]interface{}) []*models.Chunk {
	var chunks []*models.Chunk
	start := 0
	idx := 0

	for start < len(text) {
		end := start + c.ChunkSize
		if end > len(text) {
			end = len(text)
		}

		chunkText := text[start:end]
		chunks = append(chunks, &models.Chunk{
			DocumentID: docID,
			Content:    chunkText,
			ChunkIndex: idx,
			TokenCount: len(strings.Split(chunkText, " ")),
			Metadata:   meta,
		})

		idx++
		start = end - c.ChunkOverlap
		if start < 0 {
			start = 0
		}
	}

	return chunks
}

// ── Sentence-aware chunking ──────────────────────────────────────────────────
// Splits on sentence boundaries and groups sentences into chunks.

type SentenceChunker struct {
	MaxSentences int
	ChunkOverlap int // in sentences
}

func NewSentenceChunker(maxSentences, chunkOverlap int) *SentenceChunker {
	return &SentenceChunker{
		MaxSentences: maxSentences,
		ChunkOverlap: chunkOverlap,
	}
}

func (c *SentenceChunker) Split(text string, docID string, meta map[string]interface{}) []*models.Chunk {
	// Go's regexp (RE2) does not support lookbehind assertions, so we use
	// FindAllStringIndex to locate sentence boundaries manually.
	re := regexp.MustCompile(`[.!?]\s+`)
	matches := re.FindAllStringIndex(text, -1)

	sentences := make([]string, 0, len(matches)+1)
	lastEnd := 0
	for _, m := range matches {
		start := m[0] // position of the punctuation character
		end := m[1]   // position after the trailing whitespace
		// Include the punctuation but not the whitespace
		sentences = append(sentences, text[lastEnd:start+1])
		lastEnd = end
	}
	// Append any remaining text after the last sentence boundary
	if lastEnd < len(text) {
		sentences = append(sentences, text[lastEnd:])
	}
	var chunks []*models.Chunk
	idx := 0
	i := 0

	for i < len(sentences) {
		var buf strings.Builder
		count := 0
		start := i

		for i < len(sentences) && count < c.MaxSentences {
			buf.WriteString(sentences[i])
			buf.WriteString(" ")
			count++
			i++
		}

		chunks = append(chunks, &models.Chunk{
			DocumentID: docID,
			Content:    strings.TrimSpace(buf.String()),
			ChunkIndex: idx,
			TokenCount: count,
			Metadata:   meta,
		})
		idx++
		i = start + (c.MaxSentences - c.ChunkOverlap)
		if i < start+1 {
			i = start + 1
		}
	}

	return chunks
}

// ── Paragraph chunking ───────────────────────────────────────────────────────
// Each paragraph (split by double newline) becomes a chunk.

type ParagraphChunker struct{}

func NewParagraphChunker() *ParagraphChunker {
	return &ParagraphChunker{}
}

func (c *ParagraphChunker) Split(text string, docID string, meta map[string]interface{}) []*models.Chunk {
	paragraphs := strings.Split(text, "\n\n")
	var chunks []*models.Chunk

	for idx, para := range paragraphs {
		trimmed := strings.TrimSpace(para)
		if trimmed == "" {
			continue
		}
		chunks = append(chunks, &models.Chunk{
			DocumentID: docID,
			Content:    trimmed,
			ChunkIndex: idx,
			TokenCount: len(strings.Split(trimmed, " ")),
			Metadata:   meta,
		})
	}

	return chunks
}
