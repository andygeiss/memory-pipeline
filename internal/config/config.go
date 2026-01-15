package config

import (
	"os"
	"strings"

	"github.com/andygeiss/cloud-native-utils/security"
)

// ConfigID is a type alias for configuration identifiers.
type ConfigID string

// Config holds the configuration parameters for the application.
type Config struct {
	MemoryDocsDir    string   `yaml:"memory_docs_dir"`
	MemoryNotesFile  string   `yaml:"memory_notes_file"`
	MemorySourceDir  string   `yaml:"memory_source_dir"`
	MemoryStateFile  string   `yaml:"memory_state_file"`
	OpenAIAPIKey     string   `yaml:"openai_api_key"`
	OpenAIBaseURL    string   `yaml:"openai_base_url"`
	OpenAIChatModel  string   `yaml:"openai_chat_model"`
	OpenAIEmbedModel string   `yaml:"openai_embed_model"`
	FileExtensions   []string `yaml:"file_extensions"`
}

// NewConfig creates a new Config instance with default values.
func NewConfig() Config {
	// Get file extensions from environment variable or use defaults.
	exts := strings.Split(os.Getenv("APP_FILE_EXTENSIONS"), ",")
	if len(exts) == 1 && exts[0] == "" {
		exts = []string{".md", ".txt", ".go"}
	}

	return Config{
		FileExtensions:   exts,
		MemoryDocsDir:    security.ParseStringOrDefault(os.Getenv("MEMORY_DOCS_DIR"), "docs"),
		MemoryNotesFile:  security.ParseStringOrDefault(os.Getenv("MEMORY_FILE"), ".memory-notes.json"),
		MemorySourceDir:  security.ParseStringOrDefault(os.Getenv("MEMORY_SOURCE_DIR"), "."),
		MemoryStateFile:  security.ParseStringOrDefault(os.Getenv("MEMORY_STATE_FILE"), ".memory-state.json"),
		OpenAIAPIKey:     security.ParseStringOrDefault(os.Getenv("OPENAI_API_KEY"), "not-used-in-local-llm-mode"),
		OpenAIBaseURL:    security.ParseStringOrDefault(os.Getenv("OPENAI_BASE_URL"), "http://localhost:1234/v1"),
		OpenAIChatModel:  security.ParseStringOrDefault(os.Getenv("OPENAI_CHAT_MODEL"), "qwen/qwen3-coder-30b"),
		OpenAIEmbedModel: security.ParseStringOrDefault(os.Getenv("OPENAI_EMBED_MODEL"), "text-embedding-qwen3-embedding-0.6b"),
	}
}
