package documents

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"
)

// DocumentProcessor handles file parsing, text extraction, and format conversion
// for various document types (PDF, Markdown, HTML, plain text, etc.).
type DocumentProcessor struct {
	MaxFileSize int // bytes
	AllowedExts []string
}

func NewDocumentProcessor(maxFileSize int) *DocumentProcessor {
	return &DocumentProcessor{
		MaxFileSize: maxFileSize,
		AllowedExts: []string{"pdf", "md", "markdown", "html", "htm", "txt", "json", "csv"},
	}
}

// ProcessResult contains the extracted text and metadata from a raw file.
type ProcessResult struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Format   string `json:"format"` // "pdf", "markdown", "html", "text"
	Size     int    `json:"size"`
	Checksum string `json:"checksum"`
}

// ProcessFile extracts text content from a raw byte stream based on file type.
func (dp *DocumentProcessor) ProcessFile(filename string, data []byte) *ProcessResult {
	ext := dp.DetectFormat(filename)

	if !dp.IsAllowed(ext) {
		log.Printf("DocumentProcessor: Unsupported file type: %s", ext)
		return nil
	}

	if len(data) > dp.MaxFileSize {
		log.Printf("DocumentProcessor: File exceeds size limit: %s", filename)
		return nil
	}

	var content string
	switch ext {
	case "pdf":
		content = dp.ExtractPDF(data)
	case "md", "markdown":
		content = dp.ExtractMarkdown(data)
	case "html", "htm":
		content = dp.ExtractHTML(data)
	default:
		content = string(data)
	}

	return &ProcessResult{
		Title:    filename,
		Content:  content,
		Format:   ext,
		Size:     len(data),
		Checksum: dp.ComputeChecksum(data),
	}
}

// DetectFormat returns the file extension in lowercase.
func (dp *DocumentProcessor) DetectFormat(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return "txt"
	}
	return strings.ToLower(parts[len(parts)-1])
}

// IsAllowed checks whether the file extension is in the allowed list.
func (dp *DocumentProcessor) IsAllowed(ext string) bool {
	for _, allowed := range dp.AllowedExts {
		if allowed == ext {
			return true
		}
	}
	return false
}

// ComputeChecksum returns a SHA-256 hex digest of the data.
func (dp *DocumentProcessor) ComputeChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// ExtractPDF extracts text content from a PDF byte stream using the ledongthuc/pdf library.
func (dp *DocumentProcessor) ExtractPDF(data []byte) string {
	reader := bytes.NewReader(data)
	pdfReader, err := pdf.NewReader(reader, int64(len(data)))
	if err != nil {
		log.Printf("DocumentProcessor: Failed to create PDF reader: %v", err)
		return ""
	}

	var textBuilder strings.Builder
	numPages := pdfReader.NumPage()

	for i := 1; i <= numPages; i++ {
		page := pdfReader.Page(i)

		// Get fonts used on this page for proper text extraction
		fonts := make(map[string]*pdf.Font)
		for _, fontName := range page.Fonts() {
			f := page.Font(fontName)
			fonts[fontName] = &f
		}

		pageText, err := page.GetPlainText(fonts)
		if err != nil {
			log.Printf("DocumentProcessor: Failed to extract text from page %d: %v", i, err)
			continue
		}

		textBuilder.WriteString(pageText)
		textBuilder.WriteString("\n")
	}

	extracted := strings.TrimSpace(textBuilder.String())
	if extracted == "" {
		log.Printf("DocumentProcessor: No text extracted from PDF (%d pages)", numPages)
	}

	return extracted
}

// ExtractMarkdown strips Markdown formatting and returns plain text.
func (dp *DocumentProcessor) ExtractMarkdown(data []byte) string {
	text := string(data)
	reCode := regexp.MustCompile("```[\\s\\S]*?```")
	text = reCode.ReplaceAllString(text, "")
	reHeader := regexp.MustCompile("(?m)^#+\\s*")
	text = reHeader.ReplaceAllString(text, "")
	reFormat := regexp.MustCompile("\\*\\*|__|\\*|_|~~")
	text = reFormat.ReplaceAllString(text, "")
	return strings.TrimSpace(text)
}

// ExtractHTML strips HTML tags and returns plain text.
func (dp *DocumentProcessor) ExtractHTML(data []byte) string {
	text := string(data)
	reTag := regexp.MustCompile("<[^>]*>")
	text = reTag.ReplaceAllString(text, " ")
	reSpace := regexp.MustCompile("\\s+")
	text = reSpace.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}
