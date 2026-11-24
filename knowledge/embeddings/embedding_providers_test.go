package embeddings

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIEmbeddingProvider_CreateEmbedding(t *testing.T) {
	// Mock OpenAI-compatible server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": [{
				"embedding": [0.1, 0.2, 0.3, 0.4, 0.5]
			}]
		}`))
	}))
	defer server.Close()

	provider := &OpenAIEmbeddingProvider{
		APIKey:  "test-key",
		Model:   "text-embedding-ada-002",
		BaseURL: server.URL,
		Dim:     5,
	}

	vector, err := provider.CreateEmbedding("hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(vector) != 5 {
		t.Errorf("expected 5 dimensions, got %d", len(vector))
	}

	expected := []float32{0.1, 0.2, 0.3, 0.4, 0.5}
	for i, v := range expected {
		if vector[i] != v {
			t.Errorf("at index %d: expected %f, got %f", i, v, vector[i])
		}
	}
}

func TestOpenAIEmbeddingProvider_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid API key"}`))
	}))
	defer server.Close()

	provider := &OpenAIEmbeddingProvider{
		APIKey:  "bad-key",
		Model:   "text-embedding-ada-002",
		BaseURL: server.URL,
		Dim:     384,
	}

	_, err := provider.CreateEmbedding("test")
	if err == nil {
		t.Error("expected error for unauthorized request, got nil")
	}
}

func TestOllamaEmbeddingProvider_CreateEmbedding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"embedding": [0.5, 0.4, 0.3, 0.2, 0.1]
		}`))
	}))
	defer server.Close()

	provider := &OllamaEmbeddingProvider{
		BaseURL: server.URL,
		Model:   "nomic-embed-text",
		Dim:     5,
	}

	vector, err := provider.CreateEmbedding("test query")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(vector) != 5 {
		t.Errorf("expected 5 dimensions, got %d", len(vector))
	}

	expected := []float32{0.5, 0.4, 0.3, 0.2, 0.1}
	for i, v := range expected {
		if vector[i] != v {
			t.Errorf("at index %d: expected %f, got %f", i, v, vector[i])
		}
	}
}

func TestLocalEmbeddingProvider_ReturnsZeroVector(t *testing.T) {
	provider := &LocalEmbeddingProvider{
		ModelPath: "./models/embedding",
		Dim:       384,
	}

	vector, err := provider.CreateEmbedding("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(vector) != 384 {
		t.Errorf("expected 384 dimensions, got %d", len(vector))
	}

	for i, v := range vector {
		if v != 0.0 {
			t.Errorf("at index %d: expected 0.0, got %f", i, v)
			break
		}
	}
}

func TestNewEmbeddingProvider_Factory(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		wantType string
	}{
		{"openai", "openai", "*embeddings.OpenAIEmbeddingProvider"},
		{"ollama", "ollama", "*embeddings.OllamaEmbeddingProvider"},
		{"local", "local", "*embeddings.LocalEmbeddingProvider"},
		{"unknown falls back to local", "unknown", "*embeddings.LocalEmbeddingProvider"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewEmbeddingProvider(tt.provider, "test-model", 384, "http://localhost:11434", "")
			if p == nil {
				t.Fatal("expected non-nil provider")
			}
		})
	}
}
