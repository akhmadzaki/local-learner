package config

import (
	"os"
	"strconv"
)

// Version is the application version, set at build time via -ldflags.
// Defaults to "dev" for local development.
var Version = "dev"

// ── App ───────────────────────────────────────────────────────────────────────

type AppConfig struct {
	Port    int
	Env     string
	Debug   bool
	Version string
}

// ── Database (PostgreSQL + pgvector) ─────────────────────────────────────────

type DatabaseConfig struct {
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	AutoMigrate     bool
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // in minutes
}

// ── Embeddings ────────────────────────────────────────────────────────────────

type EmbeddingsConfig struct {
	Provider string // "local" | "openai" | "ollama"
	Model    string
	Dim      int
	BaseURL  string // Ollama or OpenAI-compatible endpoint
	Device   string // "cpu" | "cuda" | "mps"
}

// ── LLM (Agent) ───────────────────────────────────────────────────────────────

type LLMConfig struct {
	Provider    string // "local" | "openai" | "anthropic"
	Model       string
	MaxTokens   int
	Temperature float64
	MaxHistory  int
	BaseURL     string
}

// ── Chunking ──────────────────────────────────────────────────────────────────

type ChunkingConfig struct {
	Strategy       string // "fixed" | "sentence" | "paragraph"
	ChunkSize      int
	ChunkOverlap   int
	MinChunkLength int
}

// ── Top-level Config ──────────────────────────────────────────────────────────

type Config struct {
	App        AppConfig
	Database   DatabaseConfig
	Embeddings EmbeddingsConfig
	LLM        LLMConfig
	Chunking   ChunkingConfig
}

func Load() Config {
	return Config{
		App: AppConfig{
			Port:    getEnvInt("APP_PORT", 8080),
			Env:     getEnv("APP_ENV", "development"),
			Debug:   getEnv("APP_DEBUG", "true") == "true",
			Version: Version,
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			Name:            getEnv("DB_NAME", "go_backend"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			AutoMigrate:     getEnv("DB_AUTO_MIGRATE", "true") == "true",
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvInt("DB_CONN_MAX_LIFETIME", 30),
		},
		Embeddings: EmbeddingsConfig{
			Provider: getEnv("EMBEDDING_PROVIDER", "local"),
			Model:    getEnv("EMBEDDING_MODEL", "qwen3-embedding:4b"),
			Dim:      getEnvInt("EMBEDDING_DIMENSION", 768),
			BaseURL:  getEnv("EMBEDDING_BASE_URL", "http://localhost:11434"),
			Device:   getEnv("EMBEDDING_DEVICE", "cpu"),
		},
		LLM: LLMConfig{
			Provider:    getEnv("LLM_PROVIDER", "local"),
			Model:       getEnv("LLM_MODEL", "llama3.2:3b"),
			MaxTokens:   getEnvInt("LLM_MAX_TOKENS", 4096),
			Temperature: getEnvFloat("LLM_TEMPERATURE", 0.1),
			MaxHistory:  getEnvInt("LLM_MAX_HISTORY", 20),
			BaseURL:     getEnv("LLM_BASE_URL", "http://localhost:11434"), // Ollama default
		},
		Chunking: ChunkingConfig{
			Strategy:       getEnv("CHUNK_STRATEGY", "sentence"),
			ChunkSize:      getEnvInt("CHUNK_SIZE", 512),
			ChunkOverlap:   getEnvInt("CHUNK_OVERLAP", 64),
			MinChunkLength: getEnvInt("CHUNK_MIN_LENGTH", 20),
		},
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return fallback
}
