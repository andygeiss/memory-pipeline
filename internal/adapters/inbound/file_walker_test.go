package inbound_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/memory-pipeline/internal/adapters/inbound"
	"github.com/andygeiss/memory-pipeline/internal/domain/extraction"
)

func TestFileWalker_New_EmptyExtensions_ReturnsError(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))
	extensions := []string{}

	// Act
	_, err := inbound.NewFileWalker(tmpDir, stateFile, extensions)

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrFileWalkerEmptyExtensions", errors.Is(err, inbound.ErrFileWalkerEmptyExtensions), true)
}

func TestFileWalker_New_EmptySourceDir_ReturnsError(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))
	sourceDir := ""

	// Act
	_, err := inbound.NewFileWalker(sourceDir, stateFile, []string{".md"})

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrFileWalkerEmptySourceDir", errors.Is(err, inbound.ErrFileWalkerEmptySourceDir), true)
}

func TestFileWalker_New_EmptyStateFile_ReturnsError(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath("")

	// Act
	_, err := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrFileWalkerEmptyStateFile", errors.Is(err, inbound.ErrFileWalkerEmptyStateFile), true)
}

func TestFileWalker_New_ValidConfig_ReturnsInstance(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))

	// Act
	fw, err := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "fw must not be nil", fw != nil, true)
}

func TestFileWalker_NextPending_EmptyDirectory_ReturnsError(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))
	fw, _ := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})

	// Act
	_, err := fw.NextPending()

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrFileStoreNoMoreFiles", errors.Is(err, extraction.ErrFileStoreNoMoreFiles), true)
}

func TestFileWalker_NextPending_WithMatchingFile_ReturnsFile(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))
	testFile := filepath.Join(tmpDir, "test.md")
	writeTestFile(t, testFile, "# Test")
	fw, _ := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})

	// Act
	file, err := fw.NextPending()

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "file must not be nil", file != nil, true)
	assert.That(t, "file status must be FilePending", file.Status, extraction.FilePending)
}

func TestFileWalker_NextPending_NoMatchingExtension_ReturnsError(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))
	testFile := filepath.Join(tmpDir, "test.txt")
	writeTestFile(t, testFile, "test")
	fw, _ := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})

	// Act
	_, err := fw.NextPending()

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrFileStoreNoMoreFiles", errors.Is(err, extraction.ErrFileStoreNoMoreFiles), true)
}

func TestFileWalker_MarkProcessing_ValidFile_UpdatesStatus(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))
	testFile := filepath.Join(tmpDir, "test.md")
	writeTestFile(t, testFile, "# Test")
	fw, _ := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})
	file, _ := fw.NextPending()

	// Act
	err := fw.MarkProcessing(file.Path)

	// Assert
	assert.That(t, "err must be nil", err, nil)
}

func TestFileWalker_MarkProcessing_UnknownFile_ReturnsError(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))
	fw, _ := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})

	// Act
	err := fw.MarkProcessing("/nonexistent/file.md")

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrFileWalkerFileNotFound", errors.Is(err, inbound.ErrFileWalkerFileNotFound), true)
}

func TestFileWalker_MarkProcessed_ValidFile_UpdatesStatus(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))
	testFile := filepath.Join(tmpDir, "test.md")
	writeTestFile(t, testFile, "# Test")
	fw, _ := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})
	file, _ := fw.NextPending()

	// Act
	err := fw.MarkProcessed(file.Path)

	// Assert
	assert.That(t, "err must be nil", err, nil)
}

func TestFileWalker_MarkProcessed_UnknownFile_ReturnsError(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))
	fw, _ := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})

	// Act
	err := fw.MarkProcessed("/nonexistent/file.md")

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrFileWalkerFileNotFound", errors.Is(err, inbound.ErrFileWalkerFileNotFound), true)
}

func TestFileWalker_MarkError_ValidFile_UpdatesStatus(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))
	testFile := filepath.Join(tmpDir, "test.md")
	writeTestFile(t, testFile, "# Test")
	fw, _ := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})
	file, _ := fw.NextPending()

	// Act
	err := fw.MarkError(file.Path, "test error reason")

	// Assert
	assert.That(t, "err must be nil", err, nil)
}

func TestFileWalker_MarkError_UnknownFile_ReturnsError(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))
	fw, _ := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})

	// Act
	err := fw.MarkError("/nonexistent/file.md", "reason")

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrFileWalkerFileNotFound", errors.Is(err, inbound.ErrFileWalkerFileNotFound), true)
}

func TestFileWalker_ReadFile_ExistingFile_ReturnsContent(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))
	testFile := filepath.Join(tmpDir, "test.md")
	expectedContent := "# Test Content\n\nThis is a test."
	writeTestFile(t, testFile, expectedContent)
	fw, _ := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})

	// Act
	content, err := fw.ReadFile(extraction.FilePath(testFile))

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "content must match", content, expectedContent)
}

func TestFileWalker_ReadFile_NonexistentFile_ReturnsError(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))
	fw, _ := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})

	// Act
	_, err := fw.ReadFile("/nonexistent/file.md")

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrFileWalkerFileNotFound", errors.Is(err, inbound.ErrFileWalkerFileNotFound), true)
}

func TestFileWalker_NextPending_MultipleExtensions_FindsMatchingFiles(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))
	writeTestFile(t, filepath.Join(tmpDir, "doc.md"), "# Doc")
	writeTestFile(t, filepath.Join(tmpDir, "notes.txt"), "Notes")
	writeTestFile(t, filepath.Join(tmpDir, "code.go"), "package main")
	fw, _ := inbound.NewFileWalker(tmpDir, stateFile, []string{".md", ".txt"})

	// Act
	file, err := fw.NextPending()

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "file must not be nil", file != nil, true)
}

func TestFileWalker_NextPending_SubdirectoryFiles_FindsFiles(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))
	subDir := filepath.Join(tmpDir, "docs")
	_ = os.MkdirAll(subDir, 0750)
	writeTestFile(t, filepath.Join(subDir, "nested.md"), "# Nested")
	fw, _ := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})

	// Act
	file, err := fw.NextPending()

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "file must not be nil", file != nil, true)
}

func TestFileWalker_NextPending_AfterProcessedAndReload_ReturnsNoPendingError(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))
	testFile := filepath.Join(tmpDir, "test.md")
	writeTestFile(t, testFile, "# Test")
	fw1, _ := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})
	file, _ := fw1.NextPending()
	_ = fw1.MarkProcessed(file.Path)
	fw2, _ := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})

	// Act
	_, err := fw2.NextPending()

	// Assert
	assert.That(t, "err must not be nil since file was already processed", err != nil, true)
	assert.That(t, "err must be ErrFileStoreNoMoreFiles", errors.Is(err, extraction.ErrFileStoreNoMoreFiles), true)
}

func TestFileWalker_NextPending_FileContentChanged_ReturnsPendingFile(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))
	testFile := filepath.Join(tmpDir, "test.md")
	writeTestFile(t, testFile, "# Original")
	fw1, _ := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})
	file, _ := fw1.NextPending()
	_ = fw1.MarkProcessed(file.Path)

	// Ensure ModTime changes (filesystem granularity can be 1s on some systems)
	time.Sleep(10 * time.Millisecond)
	writeTestFile(t, testFile, "# Modified Content")
	fw2, _ := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})

	// Act
	file, err := fw2.NextPending()

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "file must not be nil after modification", file != nil, true)
	assert.That(t, "file status must be FilePending", file.Status, extraction.FilePending)
}

func TestFileWalker_NextPending_UppercaseExtension_ReturnsFile(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))
	testFile := filepath.Join(tmpDir, "TEST.MD")
	writeTestFile(t, testFile, "# Test")
	fw, _ := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})

	// Act
	file, err := fw.NextPending()

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "file with uppercase extension must match", file != nil, true)
}

func TestFileWalker_NextPending_AfterMarkProcessed_ReturnsNoPendingError(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))
	testFile := filepath.Join(tmpDir, "test.md")
	writeTestFile(t, testFile, "# Test")
	fw, _ := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})
	file, _ := fw.NextPending()
	_ = fw.MarkProcessed(file.Path)

	// Act
	_, err := fw.NextPending()

	// Assert
	assert.That(t, "err must not be nil since no more pending files", err != nil, true)
	assert.That(t, "err must be ErrFileStoreNoMoreFiles", errors.Is(err, extraction.ErrFileStoreNoMoreFiles), true)
}

func TestFileWalker_NextPending_AfterMarkError_ReturnsNoPendingError(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	stateFile := extraction.FilePath(filepath.Join(tmpDir, "state.json"))
	testFile := filepath.Join(tmpDir, "test.md")
	writeTestFile(t, testFile, "# Test")
	fw, _ := inbound.NewFileWalker(tmpDir, stateFile, []string{".md"})
	file, _ := fw.NextPending()
	_ = fw.MarkError(file.Path, "test error")

	// Act
	_, err := fw.NextPending()

	// Assert
	assert.That(t, "err must not be nil since no more pending files", err != nil, true)
	assert.That(t, "err must be ErrFileStoreNoMoreFiles", errors.Is(err, extraction.ErrFileStoreNoMoreFiles), true)
}

// writeTestFile is a helper function that writes content to a test file with secure permissions.
func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
}
