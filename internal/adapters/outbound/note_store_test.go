package outbound_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound"
	"github.com/loopforge-ai/memory-pipeline/internal/domain/extraction"
)

func TestNoteStore_New_EmptyPath_ReturnsError(t *testing.T) {
	// Arrange
	path := ""

	// Act
	_, err := outbound.NewNoteStore(path)

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrNoteStoreEmptyPath", errors.Is(err, outbound.ErrNoteStoreEmptyPath), true)
}

func TestNoteStore_New_ValidPath_ReturnsInstance(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "notes.json")

	// Act
	ns, err := outbound.NewNoteStore(path)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "ns must not be nil", ns != nil, true)
}

func TestNoteStore_SaveNote_NewFile_CreatesFile(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "notes.json")
	ns, _ := outbound.NewNoteStore(path)
	note := createTestNote("note-1", "Test content", extraction.NoteLearning)

	// Act
	err := ns.SaveNote(note)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	_, statErr := os.Stat(path)
	assert.That(t, "file must exist", os.IsNotExist(statErr), false)
}

func TestNoteStore_SaveNote_ValidNote_PersistsContent(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "notes.json")
	ns, _ := outbound.NewNoteStore(path)
	note := createTestNote("note-1", "Test content", extraction.NotePattern)

	// Act
	err := ns.SaveNote(note)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	stored := readStoredNotes(t, path)
	assert.That(t, "stored length must be 1", len(stored), 1)
	assert.That(t, "id must be note-1", stored[0]["id"], "note-1")
	assert.That(t, "content must be Test content", stored[0]["content"], "Test content")
}

func TestNoteStore_SaveNote_MultipleNotes_PersistsAll(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "notes.json")
	ns, _ := outbound.NewNoteStore(path)
	notes := []extraction.EmbeddedNote{
		createTestNote("note-1", "First note", extraction.NoteLearning),
		createTestNote("note-2", "Second note", extraction.NoteCookbook),
	}

	// Act
	for _, note := range notes {
		_ = ns.SaveNote(note)
	}

	// Assert
	stored := readStoredNotes(t, path)
	assert.That(t, "stored length must be 2", len(stored), 2)
}

func TestNoteStore_SaveNote_ExistingNote_UpdatesContent(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "notes.json")
	ns, _ := outbound.NewNoteStore(path)
	note := createTestNote("note-1", "Original content", extraction.NoteLearning)
	_ = ns.SaveNote(note)
	note.Note.Content = "Updated content"
	note.Embedding = []float32{0.5, 0.6}

	// Act
	err := ns.SaveNote(note)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	stored := readStoredNotes(t, path)
	assert.That(t, "stored length must be 1 after update", len(stored), 1)
	assert.That(t, "content must be updated", stored[0]["content"], "Updated content")
}

func TestNoteStore_New_ExistingFile_LoadsNotes(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "notes.json")
	ns1, _ := outbound.NewNoteStore(path)
	note := createTestNote("note-1", "Persistent content", extraction.NoteDecision)
	_ = ns1.SaveNote(note)
	ns2, _ := outbound.NewNoteStore(path)
	note2 := createTestNote("note-2", "Second note", extraction.NoteLearning)

	// Act
	err := ns2.SaveNote(note2)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	stored := readStoredNotes(t, path)
	assert.That(t, "stored length must be 2 after loading existing", len(stored), 2)
}

func TestNoteStore_SaveNote_NestedPath_CreatesDirectory(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "subdir", "nested", "notes.json")
	ns, _ := outbound.NewNoteStore(path)
	note := createTestNote("note-1", "Test content", extraction.NoteLearning)

	// Act
	err := ns.SaveNote(note)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	_, statErr := os.Stat(path)
	assert.That(t, "file must exist in nested directory", os.IsNotExist(statErr), false)
}

func TestNoteStore_SaveNote_WithEmbedding_PreservesEmbedding(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "notes.json")
	ns, _ := outbound.NewNoteStore(path)
	expectedEmbedding := []float32{0.123, 0.456, 0.789, 1.0, -0.5}
	note := extraction.EmbeddedNote{
		Note: extraction.MemoryNote{
			ID:      "note-1",
			Content: "Test content",
			Kind:    extraction.NoteLearning,
			Path:    "/path/to/file.md",
		},
		Embedding: expectedEmbedding,
	}

	// Act
	err := ns.SaveNote(note)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	data, _ := os.ReadFile(path) //nolint:gosec // Test reads from controlled test paths
	var stored []struct {
		Embedding []float32 `json:"embedding"`
	}
	_ = json.Unmarshal(data, &stored)
	assert.That(t, "stored length must be 1", len(stored), 1)
	assert.That(t, "embedding length must match", len(stored[0].Embedding), len(expectedEmbedding))
	for i, v := range expectedEmbedding {
		assert.That(t, "embedding value must match", stored[0].Embedding[i], v)
	}
}

func TestNoteStore_SaveNote_AllKinds_PreservesAllKinds(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "notes.json")
	ns, _ := outbound.NewNoteStore(path)
	kinds := []extraction.NoteKind{
		extraction.NoteLearning,
		extraction.NotePattern,
		extraction.NoteCookbook,
		extraction.NoteDecision,
	}

	// Act
	for i, kind := range kinds {
		note := createTestNote(extraction.NodeID(string(rune('a'+i))), "Content", kind)
		_ = ns.SaveNote(note)
	}

	// Assert
	stored := readStoredNotes(t, path)
	assert.That(t, "stored length must match kinds length", len(stored), len(kinds))
}

// createTestNote is a helper function that creates an EmbeddedNote for testing.
func createTestNote(id extraction.NodeID, content string, kind extraction.NoteKind) extraction.EmbeddedNote {
	return extraction.EmbeddedNote{
		Note: extraction.MemoryNote{
			ID:      id,
			Content: extraction.NoteContent(content),
			Kind:    kind,
			Path:    "/path/to/file.md",
		},
		Embedding: []float32{0.1, 0.2, 0.3},
	}
}

// readStoredNotes is a helper function that reads and unmarshals the stored notes.
func readStoredNotes(t *testing.T, path string) []map[string]any {
	t.Helper()

	data, err := os.ReadFile(path) //nolint:gosec // Test helper reads from controlled test paths
	if err != nil {
		t.Fatalf("failed to read notes file: %v", err)
	}

	var stored []map[string]any
	if err := json.Unmarshal(data, &stored); err != nil {
		t.Fatalf("failed to unmarshal notes: %v", err)
	}

	return stored
}
