package outbound_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/memory-pipeline/internal/adapters/outbound"
	"github.com/andygeiss/memory-pipeline/internal/domain/extraction"
)

func TestNewMarkdownWriter_EmptyPath_ReturnsError(t *testing.T) {
	// Act
	_, err := outbound.NewMarkdownWriter("")

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrMarkdownWriterEmptyPath", errors.Is(err, outbound.ErrMarkdownWriterEmptyPath), true)
}

func TestNewMarkdownWriter_ValidPath_ReturnsInstance(t *testing.T) {
	// Act
	mw, err := outbound.NewMarkdownWriter("/tmp/test-docs")

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "mw must not be nil", mw != nil, true)
}

func TestMarkdownWriter_WriteDoc_CollectsNotes(t *testing.T) {
	// Arrange
	mw, _ := outbound.NewMarkdownWriter("/tmp/test-docs")
	note := extraction.MemoryNote{
		Content: "Test content",
		ID:      "test-id",
		Kind:    extraction.NoteLearning,
		Path:    "/test/file.go",
	}

	// Act
	err := mw.WriteDoc(note)

	// Assert
	assert.That(t, "err must be nil", err, nil)
}

func TestMarkdownWriter_Finalize_CreatesFiles(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	mw, _ := outbound.NewMarkdownWriter(tmpDir)
	notes := []extraction.MemoryNote{
		{ID: "1", Content: "Learning content", Kind: extraction.NoteLearning, Path: "/test/a.go"},
		{ID: "2", Content: "Pattern content", Kind: extraction.NotePattern, Path: "/test/b.go"},
		{ID: "3", Content: "Cookbook content", Kind: extraction.NoteCookbook, Path: "/test/c.go"},
		{ID: "4", Content: "Decision content", Kind: extraction.NoteDecision, Path: "/test/d.go"},
	}

	for _, note := range notes {
		_ = mw.WriteDoc(note)
	}

	// Act
	err := mw.Finalize()

	// Assert
	assert.That(t, "err must be nil", err, nil)

	// Verify files exist
	expectedFiles := []string{"index.md", "learnings.md", "patterns.md", "cookbooks.md", "decisions.md"}
	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		_, err := os.Stat(path)
		assert.That(t, file+" must exist", err == nil, true)
	}
}

func TestMarkdownWriter_Finalize_IndexContainsLinks(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	mw, _ := outbound.NewMarkdownWriter(tmpDir)
	note := extraction.MemoryNote{
		Content: "Test content",
		ID:      "1",
		Kind:    extraction.NoteLearning,
		Path:    "/test/file.go",
	}
	_ = mw.WriteDoc(note)

	// Act
	err := mw.Finalize()

	// Assert
	assert.That(t, "err must be nil", err, nil)

	content, _ := os.ReadFile(filepath.Clean(filepath.Join(tmpDir, "index.md")))
	assert.That(t, "index must contain learnings link", strings.Contains(string(content), "[Learnings](learnings.md)"), true)
	assert.That(t, "index must contain patterns link", strings.Contains(string(content), "[Patterns](patterns.md)"), true)
	assert.That(t, "index must contain cookbooks link", strings.Contains(string(content), "[Cookbooks](cookbooks.md)"), true)
	assert.That(t, "index must contain decisions link", strings.Contains(string(content), "[Decisions](decisions.md)"), true)
}

func TestMarkdownWriter_Finalize_CategoryFileContainsNotes(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	mw, _ := outbound.NewMarkdownWriter(tmpDir)
	note := extraction.MemoryNote{
		Content: "Important learning about Go",
		ID:      "1",
		Kind:    extraction.NoteLearning,
		Path:    "/test/file.go",
	}
	_ = mw.WriteDoc(note)

	// Act
	err := mw.Finalize()

	// Assert
	assert.That(t, "err must be nil", err, nil)

	content, _ := os.ReadFile(filepath.Clean(filepath.Join(tmpDir, "learnings.md")))
	assert.That(t, "learnings must contain note content", strings.Contains(string(content), "Important learning about Go"), true)
	assert.That(t, "learnings must contain file path", strings.Contains(string(content), "/test/file.go"), true)
}

func TestMarkdownWriter_Finalize_EmptyCategory_ShowsPlaceholder(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	mw, _ := outbound.NewMarkdownWriter(tmpDir)
	// Don't add any notes - all categories should be empty

	// Act
	err := mw.Finalize()

	// Assert
	assert.That(t, "err must be nil", err, nil)

	content, _ := os.ReadFile(filepath.Clean(filepath.Join(tmpDir, "learnings.md")))
	assert.That(t, "empty category must show placeholder", strings.Contains(string(content), "No notes in this category yet"), true)
}

func TestMarkdownWriter_Finalize_GroupsByFilePath(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	mw, _ := outbound.NewMarkdownWriter(tmpDir)
	notes := []extraction.MemoryNote{
		{ID: "1", Content: "First note", Kind: extraction.NoteLearning, Path: "/test/alpha.go"},
		{ID: "2", Content: "Second note", Kind: extraction.NoteLearning, Path: "/test/beta.go"},
		{ID: "3", Content: "Third note", Kind: extraction.NoteLearning, Path: "/test/alpha.go"},
	}
	for _, note := range notes {
		_ = mw.WriteDoc(note)
	}

	// Act
	err := mw.Finalize()

	// Assert
	assert.That(t, "err must be nil", err, nil)

	content, _ := os.ReadFile(filepath.Clean(filepath.Join(tmpDir, "learnings.md")))
	// Both file paths should be present as headers
	assert.That(t, "learnings must contain alpha.go header", strings.Contains(string(content), "## /test/alpha.go"), true)
	assert.That(t, "learnings must contain beta.go header", strings.Contains(string(content), "## /test/beta.go"), true)
}
