package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andygeiss/memory-pipeline/internal/adapters/inbound"
	"github.com/andygeiss/memory-pipeline/internal/domain/extraction"
)

// BenchmarkFileWalker benchmarks the file walker scanning and state management.
// This exercises the hot path for file discovery and hash computation.
func BenchmarkFileWalker(b *testing.B) {
	tmpDir := b.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))

	// Create test files
	for i := range 10 {
		content := []byte("# Test File\n\nThis is test content for benchmarking.\n")
		filename := filepath.Join(tmpDir, "test"+string(rune('0'+i))+".md")
		if err := os.WriteFile(filename, content, 0600); err != nil {
			b.Fatal(err)
		}
	}

	for b.Loop() {
		fw, err := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})
		if err != nil {
			b.Fatal(err)
		}

		// Simulate the file discovery loop
		for {
			file, err := fw.NextPending()
			if err != nil {
				break
			}
			_ = fw.MarkProcessing(file.Path)
			_, _ = fw.ReadFile(file.Path)
			_ = fw.MarkProcessed(file.Path)
		}

		// Clean up state file for next iteration
		_ = os.Remove(string(stateFile))
	}
}

// BenchmarkFileWalkerScan benchmarks just the directory scanning.
func BenchmarkFileWalkerScan(b *testing.B) {
	tmpDir := b.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))

	// Create nested directory structure with files
	for i := range 5 {
		subDir := filepath.Join(tmpDir, "subdir"+string(rune('0'+i)))
		if err := os.MkdirAll(subDir, 0755); err != nil {
			b.Fatal(err)
		}
		for j := range 5 {
			content := []byte("# Nested Test\n\nContent for nested file.\n")
			filename := filepath.Join(subDir, "file"+string(rune('0'+j))+".md")
			if err := os.WriteFile(filename, content, 0600); err != nil {
				b.Fatal(err)
			}
		}
	}

	for b.Loop() {
		fw, err := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})
		if err != nil {
			b.Fatal(err)
		}

		// Just scan for pending files
		_, _ = fw.NextPending()

		// Clean up state file
		_ = os.Remove(string(stateFile))
	}
}

// BenchmarkServiceConfig benchmarks service configuration validation.
func BenchmarkServiceConfig(b *testing.B) {
	cfg := extraction.ServiceConfig{
		Embeddings: &mockEmbeddingClient{},
		Files:      &mockFileStore{},
		LLM:        &mockLLMClient{},
		Notes:      &mockNoteStore{},
	}

	for b.Loop() {
		_, _ = extraction.NewService(cfg)
	}
}

// BenchmarkServiceRun benchmarks the full service run with mocked adapters.
func BenchmarkServiceRun(b *testing.B) {
	for b.Loop() {
		svc, _ := extraction.NewService(extraction.ServiceConfig{
			Embeddings: &mockEmbeddingClient{},
			Files:      &mockFileStore{fileCount: 10},
			LLM:        &mockLLMClient{},
			Notes:      &mockNoteStore{},
		})
		_ = svc.Run()
	}
}

// Mock implementations for benchmarking

type mockFileStore struct {
	fileCount int
	current   int
}

func (m *mockFileStore) NextPending() (*extraction.File, error) {
	if m.current >= m.fileCount {
		return nil, extraction.ErrFileStoreNoMoreFiles
	}
	m.current++
	return &extraction.File{
		Hash:   extraction.FileHash("hash" + string(rune('0'+m.current))),
		Path:   extraction.FilePath("/test/file" + string(rune('0'+m.current)) + ".md"),
		Status: extraction.FilePending,
	}, nil
}

func (m *mockFileStore) ReadFile(_ extraction.FilePath) (string, error) {
	return "# Test\n\nThis is test content for extraction.", nil
}

func (m *mockFileStore) MarkProcessing(_ extraction.FilePath) error { return nil }
func (m *mockFileStore) MarkProcessed(_ extraction.FilePath) error  { return nil }
func (m *mockFileStore) MarkError(_ extraction.FilePath, _ string) error {
	return nil
}

type mockLLMClient struct{}

func (m *mockLLMClient) ExtractNotes(path extraction.FilePath, _ string) ([]extraction.MemoryNote, error) {
	return []extraction.MemoryNote{
		{
			ID:      extraction.NodeID("note-1"),
			Content: extraction.NoteContent("Test learning note"),
			Kind:    extraction.NoteLearning,
			Path:    path,
		},
		{
			ID:      extraction.NodeID("note-2"),
			Content: extraction.NoteContent("Test pattern note"),
			Kind:    extraction.NotePattern,
			Path:    path,
		},
	}, nil
}

type mockEmbeddingClient struct{}

func (m *mockEmbeddingClient) Embed(note extraction.MemoryNote) (extraction.EmbeddedNote, error) {
	return extraction.EmbeddedNote{
		Note:      note,
		Embedding: make([]float32, 384), // typical embedding size
	}, nil
}

type mockNoteStore struct{}

func (m *mockNoteStore) SaveNote(_ extraction.EmbeddedNote) error { return nil }
