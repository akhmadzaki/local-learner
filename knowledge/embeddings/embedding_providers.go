package embeddings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// EmbeddingProvider is the interface for all embedding model backends.
type EmbeddingProvider interface {
	CreateEmbedding(text string) ([]float32, error)
	ModelName() string
	Dimension() int
}

// ── OpenAI-compatible provider ───────────────────────────────────────────────

type OpenAIEmbeddingProvider struct {
	APIKey  string
	Model   string
	BaseURL string
	Dim     int
}

func (p *OpenAIEmbeddingProvider) CreateEmbedding(text string) ([]float32, error) {
	body := map[string]interface{}{
		"input": text,
		"model": p.Model,
	}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", p.BaseURL+"/v1/embeddings", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		var errBody map[string]interface{}
		if decodeErr := json.NewDecoder(resp.Body).Decode(&errBody); decodeErr == nil {
			if msg, ok := errBody["error"].(string); ok {
				return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, msg)
			}
		}
		return nil, fmt.Errorf("OpenAI API error: status %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	rawData, ok := data["data"].([]interface{})
	if !ok || len(rawData) == 0 {
		return nil, err
	}

	rawEmbedding, ok := rawData[0].(map[string]interface{})["embedding"].([]interface{})
	if !ok {
		return nil, err
	}

	vector := make([]float32, len(rawEmbedding))
	for i, v := range rawEmbedding {
		vector[i] = float32(v.(float64))
	}
	return vector, nil
}

func (p *OpenAIEmbeddingProvider) ModelName() string {
	return p.Model
}

func (p *OpenAIEmbeddingProvider) Dimension() int {
	return p.Dim
}

// ── Local / on-device provider (stub) ────────────────────────────────────────

type LocalEmbeddingProvider struct {
	ModelPath string
	Dim       int
}

func (p *LocalEmbeddingProvider) CreateEmbedding(text string) ([]float32, error) {
	log.Printf("LocalEmbeddingProvider: stub returning zero vector for %d dimensions", p.Dim)
	vector := make([]float32, p.Dim)
	return vector, nil
}

func (p *LocalEmbeddingProvider) ModelName() string {
	return "local-stub"
}

func (p *LocalEmbeddingProvider) Dimension() int {
	return p.Dim
}

// ── Ollama embedding provider ────────────────────────────────────────────────

// OllamaEmbeddingProvider generates embeddings using Ollama's /api/embeddings endpoint.
// Supports models like nomic-embed-text, all-minilm, etc.
type OllamaEmbeddingProvider struct {
	BaseURL string
	Model   string
	Dim     int
}

func (p *OllamaEmbeddingProvider) CreateEmbedding(text string) ([]float32, error) {
	body := map[string]interface{}{
		"model":  p.Model,
		"prompt": text,
	}
	jsonBody, _ := json.Marshal(body)

	resp, err := http.Post(p.BaseURL+"/api/embeddings", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	// Ollama returns embedding as []interface{} under the "embedding" key
	embedding, ok := data["embedding"].([]interface{})
	if !ok {
		return nil, err
	}

	vector := make([]float32, len(embedding))
	for i, v := range embedding {
		vector[i] = float32(v.(float64))
	}
	return vector, nil
}

func (p *OllamaEmbeddingProvider) ModelName() string {
	return "ollama:" + p.Model
}

func (p *OllamaEmbeddingProvider) Dimension() int {
	return p.Dim
}

// ── Factory ──────────────────────────────────────────────────────────────────

func NewEmbeddingProvider(provider string, model string, dim int, baseURL string, apiKey string) EmbeddingProvider {
	switch provider {
	case "openai":
		return &OpenAIEmbeddingProvider{
			APIKey:  apiKey,
			Model:   model,
			BaseURL: baseURL,
			Dim:     dim,
		}
	case "ollama":
		return &OllamaEmbeddingProvider{
			BaseURL: baseURL,
			Model:   model,
			Dim:     dim,
		}
	case "local":
		return &LocalEmbeddingProvider{
			ModelPath: "./models/embedding",
			Dim:       dim,
		}
	default:
		log.Printf("EmbeddingProvider: unknown provider '%s', falling back to local", provider)
		return &LocalEmbeddingProvider{
			ModelPath: "./models/embedding",
			Dim:       dim,
		}
	}
}
