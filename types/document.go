package types

// DocumentType enumerates the supported document formats.
type DocumentType string

const (
	DocTypePDF       DocumentType = "pdf"
	DocTypeMarkdown  DocumentType = "markdown"
	DocTypeHTML      DocumentType = "html"
	DocTypePlainText DocumentType = "text"
	DocTypeJSON      DocumentType = "json"
	DocTypeCSV       DocumentType = "csv"
)

// DocumentStatus represents the lifecycle state of a document.
type DocumentStatus string

const (
	DocStatusPending    DocumentStatus = "pending"
	DocStatusProcessing DocumentStatus = "processing"
	DocStatusIndexed    DocumentStatus = "indexed"
	DocStatusFailed     DocumentStatus = "failed"
	DocStatusDeleted    DocumentStatus = "deleted"
)

// DocumentMetadata is a typed structure for common metadata fields.
type DocumentMetadata struct {
	Author    string   `json:"author"`
	SourceURL string   `json:"source_url"`
	Tags      []string `json:"tags"`
	Category  string   `json:"category"`
	Language  string   `json:"language"`
	FileSize  int      `json:"file_size"`
	FileType  string   `json:"file_type"`
}
