package extraction_test

import (
	"errors"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/memory-pipeline/internal/domain/extraction"
)

const testFileContent = "Test content"

// === Mock Implementations ===

// mockEmbeddingClient implements extraction.EmbeddingClient for testing.
type mockEmbeddingClient struct {
	embedFunc func(note extraction.MemoryNote) (extraction.EmbeddedNote, error)
	calls     []extraction.MemoryNote
}

func (m *mockEmbeddingClient) Embed(note extraction.MemoryNote) (extraction.EmbeddedNote, error) {
	m.calls = append(m.calls, note)
	if m.embedFunc != nil {
		return m.embedFunc(note)
	}
	return extraction.EmbeddedNote{
		Embedding: []float32{0.1, 0.2, 0.3},
		Note:      note,
	}, nil
}

// mockFileStore implements extraction.FileStore for testing.
type mockFileStore struct {
	fileContents    map[extraction.FilePath]string
	markErrorFunc   func(path extraction.FilePath, reason string) error
	markProcessFunc func(path extraction.FilePath) error
	files           []extraction.File
	processingPaths []extraction.FilePath
	processedPaths  []extraction.FilePath
	errorPaths      []extraction.FilePath
	nextIndex       int
}

func newMockFileStore() *mockFileStore {
	return &mockFileStore{
		fileContents: make(map[extraction.FilePath]string),
	}
}

func (m *mockFileStore) MarkError(path extraction.FilePath, reason string) error {
	m.errorPaths = append(m.errorPaths, path)
	if m.markErrorFunc != nil {
		return m.markErrorFunc(path, reason)
	}
	return nil
}

func (m *mockFileStore) MarkProcessed(path extraction.FilePath) error {
	m.processedPaths = append(m.processedPaths, path)
	if m.markProcessFunc != nil {
		return m.markProcessFunc(path)
	}
	return nil
}

func (m *mockFileStore) MarkProcessing(path extraction.FilePath) error {
	m.processingPaths = append(m.processingPaths, path)
	return nil
}

func (m *mockFileStore) NextPending() (*extraction.File, error) {
	if m.nextIndex >= len(m.files) {
		return nil, extraction.ErrFileStoreNoMoreFiles
	}
	file := m.files[m.nextIndex]
	m.nextIndex++
	return &file, nil
}

func (m *mockFileStore) ReadFile(path extraction.FilePath) (string, error) {
	content, ok := m.fileContents[path]
	if !ok {
		return "", errors.New("file not found")
	}
	return content, nil
}

// mockLLMClient implements extraction.LLMClient for testing.
type mockLLMClient struct {
	extractFunc func(filePath extraction.FilePath, contents string) ([]extraction.MemoryNote, error)
	calls       []string
}

func (m *mockLLMClient) ExtractNotes(filePath extraction.FilePath, contents string) ([]extraction.MemoryNote, error) {
	m.calls = append(m.calls, contents)
	if m.extractFunc != nil {
		return m.extractFunc(filePath, contents)
	}
	return []extraction.MemoryNote{
		{
			Content: extraction.NoteContent("Extracted note from " + string(filePath)),
			ID:      "note-1",
			Kind:    extraction.NoteLearning,
			Path:    filePath,
		},
	}, nil
}

// mockNoteStore implements extraction.NoteStore for testing.
type mockNoteStore struct {
	saveFunc func(note extraction.EmbeddedNote) error
	notes    []extraction.EmbeddedNote
}

func (m *mockNoteStore) SaveNote(note extraction.EmbeddedNote) error {
	m.notes = append(m.notes, note)
	if m.saveFunc != nil {
		return m.saveFunc(note)
	}
	return nil
}

// noOpProgress is a no-op progress function for testing.
func noOpProgress(current, total int, desc string) {}

// === ServiceConfig Tests ===

func TestServiceConfig_Validate_MissingEmbeddings_ReturnsError(t *testing.T) {
	// Arrange
	cfg := extraction.ServiceConfig{
		Embeddings: nil,
		Files:      newMockFileStore(),
		LLM:        &mockLLMClient{},
		Notes:      &mockNoteStore{},
		ProgressFn: noOpProgress,
	}

	// Act
	err := cfg.Validate()

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrServiceConfigMissingEmbeddingClient", errors.Is(err, extraction.ErrServiceConfigMissingEmbeddingClient), true)
}

func TestServiceConfig_Validate_MissingFiles_ReturnsError(t *testing.T) {
	// Arrange
	cfg := extraction.ServiceConfig{
		Embeddings: &mockEmbeddingClient{},
		Files:      nil,
		LLM:        &mockLLMClient{},
		Notes:      &mockNoteStore{},
		ProgressFn: noOpProgress,
	}

	// Act
	err := cfg.Validate()

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrServiceConfigMissingFileStore", errors.Is(err, extraction.ErrServiceConfigMissingFileStore), true)
}

func TestServiceConfig_Validate_MissingLLM_ReturnsError(t *testing.T) {
	// Arrange
	cfg := extraction.ServiceConfig{
		Embeddings: &mockEmbeddingClient{},
		Files:      newMockFileStore(),
		LLM:        nil,
		Notes:      &mockNoteStore{},
		ProgressFn: noOpProgress,
	}

	// Act
	err := cfg.Validate()

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrServiceConfigMissingLLMClient", errors.Is(err, extraction.ErrServiceConfigMissingLLMClient), true)
}

func TestServiceConfig_Validate_MissingNotes_ReturnsError(t *testing.T) {
	// Arrange
	cfg := extraction.ServiceConfig{
		Embeddings: &mockEmbeddingClient{},
		Files:      newMockFileStore(),
		LLM:        &mockLLMClient{},
		Notes:      nil,
		ProgressFn: noOpProgress,
	}

	// Act
	err := cfg.Validate()

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrServiceConfigMissingNoteStore", errors.Is(err, extraction.ErrServiceConfigMissingNoteStore), true)
}

func TestServiceConfig_Validate_MissingProgressFn_ReturnsError(t *testing.T) {
	// Arrange
	cfg := extraction.ServiceConfig{
		Embeddings: &mockEmbeddingClient{},
		Files:      newMockFileStore(),
		LLM:        &mockLLMClient{},
		Notes:      &mockNoteStore{},
		ProgressFn: nil,
	}

	// Act
	err := cfg.Validate()

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrServiceConfigMissingProgressBar", errors.Is(err, extraction.ErrServiceConfigMissingProgressBar), true)
}

func TestServiceConfig_Validate_AllPresent_ReturnsNil(t *testing.T) {
	// Arrange
	cfg := extraction.ServiceConfig{
		Embeddings: &mockEmbeddingClient{},
		Files:      newMockFileStore(),
		LLM:        &mockLLMClient{},
		Notes:      &mockNoteStore{},
		ProgressFn: noOpProgress,
	}

	// Act
	err := cfg.Validate()

	// Assert
	assert.That(t, "err must be nil", err, nil)
}

// === NewService Tests ===

func TestService_New_InvalidConfig_ReturnsError(t *testing.T) {
	// Arrange
	cfg := extraction.ServiceConfig{
		Embeddings: nil,
		Files:      newMockFileStore(),
		LLM:        &mockLLMClient{},
		Notes:      &mockNoteStore{},
		ProgressFn: noOpProgress,
	}

	// Act
	_, err := extraction.NewService(cfg)

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
}

func TestService_New_ValidConfig_ReturnsInstance(t *testing.T) {
	// Arrange
	cfg := extraction.ServiceConfig{
		Embeddings: &mockEmbeddingClient{},
		Files:      newMockFileStore(),
		LLM:        &mockLLMClient{},
		Notes:      &mockNoteStore{},
		ProgressFn: noOpProgress,
	}

	// Act
	svc, err := extraction.NewService(cfg)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "svc must not be nil", svc != nil, true)
}

// === Service.Run Tests ===

func TestService_Run_NoPendingFiles_ReturnsNil(t *testing.T) {
	// Arrange
	fs := newMockFileStore()
	svc, _ := extraction.NewService(extraction.ServiceConfig{
		Embeddings: &mockEmbeddingClient{},
		Files:      fs,
		LLM:        &mockLLMClient{},
		Notes:      &mockNoteStore{},
		ProgressFn: noOpProgress,
	})

	// Act
	err := svc.Run()

	// Assert
	assert.That(t, "err must be nil", err, nil)
}

func TestService_Run_SingleFile_ProcessesSuccessfully(t *testing.T) {
	// Arrange
	fs := newMockFileStore()
	fs.files = []extraction.File{
		{Hash: "hash1", Path: "/test/file1.md", Status: extraction.FilePending},
	}
	fs.fileContents["/test/file1.md"] = testFileContent
	ns := &mockNoteStore{}
	ec := &mockEmbeddingClient{}
	svc, _ := extraction.NewService(extraction.ServiceConfig{
		Embeddings: ec,
		Files:      fs,
		LLM:        &mockLLMClient{},
		Notes:      ns,
		ProgressFn: noOpProgress,
	})

	// Act
	err := svc.Run()

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "processing paths length must be 1", len(fs.processingPaths), 1)
	assert.That(t, "processed paths length must be 1", len(fs.processedPaths), 1)
	assert.That(t, "saved notes length must be 1", len(ns.notes), 1)
}

func TestService_Run_MultipleFiles_ProcessesAll(t *testing.T) {
	// Arrange
	fs := newMockFileStore()
	fs.files = []extraction.File{
		{Hash: "hash1", Path: "/test/file1.md", Status: extraction.FilePending},
		{Hash: "hash2", Path: "/test/file2.md", Status: extraction.FilePending},
	}
	fs.fileContents["/test/file1.md"] = "Content 1"
	fs.fileContents["/test/file2.md"] = "Content 2"
	ns := &mockNoteStore{}
	svc, _ := extraction.NewService(extraction.ServiceConfig{
		Embeddings: &mockEmbeddingClient{},
		Files:      fs,
		LLM:        &mockLLMClient{},
		Notes:      ns,
		ProgressFn: noOpProgress,
	})

	// Act
	err := svc.Run()

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "processed paths length must be 2", len(fs.processedPaths), 2)
	assert.That(t, "saved notes length must be 2", len(ns.notes), 2)
}

func TestService_Run_FileReadError_MarksFileAsError(t *testing.T) {
	// Arrange
	fs := newMockFileStore()
	fs.files = []extraction.File{
		{Hash: "hash1", Path: "/test/missing.md", Status: extraction.FilePending},
	}
	svc, _ := extraction.NewService(extraction.ServiceConfig{
		Embeddings: &mockEmbeddingClient{},
		Files:      fs,
		LLM:        &mockLLMClient{},
		Notes:      &mockNoteStore{},
		ProgressFn: noOpProgress,
	})

	// Act
	err := svc.Run()

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "error paths length must be 1", len(fs.errorPaths), 1)
	assert.That(t, "error path must be correct", fs.errorPaths[0], extraction.FilePath("/test/missing.md"))
}

func TestService_Run_LLMError_MarksFileAsError(t *testing.T) {
	// Arrange
	fs := newMockFileStore()
	fs.files = []extraction.File{
		{Hash: "hash1", Path: "/test/file1.md", Status: extraction.FilePending},
	}
	fs.fileContents["/test/file1.md"] = testFileContent
	llm := &mockLLMClient{
		extractFunc: func(filePath extraction.FilePath, contents string) ([]extraction.MemoryNote, error) {
			return nil, errors.New("LLM error")
		},
	}
	svc, _ := extraction.NewService(extraction.ServiceConfig{
		Embeddings: &mockEmbeddingClient{},
		Files:      fs,
		LLM:        llm,
		Notes:      &mockNoteStore{},
		ProgressFn: noOpProgress,
	})

	// Act
	err := svc.Run()

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "error paths length must be 1", len(fs.errorPaths), 1)
}

func TestService_Run_EmbeddingError_ReturnsError(t *testing.T) {
	// Arrange
	fs := newMockFileStore()
	fs.files = []extraction.File{
		{Hash: "hash1", Path: "/test/file1.md", Status: extraction.FilePending},
	}
	fs.fileContents["/test/file1.md"] = testFileContent
	ec := &mockEmbeddingClient{
		embedFunc: func(note extraction.MemoryNote) (extraction.EmbeddedNote, error) {
			return extraction.EmbeddedNote{}, errors.New("embedding error")
		},
	}
	svc, _ := extraction.NewService(extraction.ServiceConfig{
		Embeddings: ec,
		Files:      fs,
		LLM:        &mockLLMClient{},
		Notes:      &mockNoteStore{},
		ProgressFn: noOpProgress,
	})

	// Act
	err := svc.Run()

	// Assert
	assert.That(t, "err must not be nil for embedding failure", err != nil, true)
}

func TestService_Run_SaveNoteError_ReturnsError(t *testing.T) {
	// Arrange
	fs := newMockFileStore()
	fs.files = []extraction.File{
		{Hash: "hash1", Path: "/test/file1.md", Status: extraction.FilePending},
	}
	fs.fileContents["/test/file1.md"] = testFileContent
	ns := &mockNoteStore{
		saveFunc: func(note extraction.EmbeddedNote) error {
			return errors.New("save error")
		},
	}
	svc, _ := extraction.NewService(extraction.ServiceConfig{
		Embeddings: &mockEmbeddingClient{},
		Files:      fs,
		LLM:        &mockLLMClient{},
		Notes:      ns,
		ProgressFn: noOpProgress,
	})

	// Act
	err := svc.Run()

	// Assert
	assert.That(t, "err must not be nil for save failure", err != nil, true)
}

func TestService_Run_MarkProcessedError_ReturnsError(t *testing.T) {
	// Arrange
	fs := newMockFileStore()
	fs.files = []extraction.File{
		{Hash: "hash1", Path: "/test/file1.md", Status: extraction.FilePending},
	}
	fs.fileContents["/test/file1.md"] = testFileContent
	fs.markProcessFunc = func(path extraction.FilePath) error {
		return errors.New("mark processed error")
	}
	svc, _ := extraction.NewService(extraction.ServiceConfig{
		Embeddings: &mockEmbeddingClient{},
		Files:      fs,
		LLM:        &mockLLMClient{},
		Notes:      &mockNoteStore{},
		ProgressFn: noOpProgress,
	})

	// Act
	err := svc.Run()

	// Assert
	assert.That(t, "err must not be nil for mark processed failure", err != nil, true)
}

func TestService_Run_NoNotesExtracted_MarksFilesProcessed(t *testing.T) {
	// Arrange
	fs := newMockFileStore()
	fs.files = []extraction.File{
		{Hash: "hash1", Path: "/test/file1.md", Status: extraction.FilePending},
	}
	fs.fileContents["/test/file1.md"] = testFileContent
	llm := &mockLLMClient{
		extractFunc: func(filePath extraction.FilePath, contents string) ([]extraction.MemoryNote, error) {
			return []extraction.MemoryNote{}, nil
		},
	}
	ns := &mockNoteStore{}
	svc, _ := extraction.NewService(extraction.ServiceConfig{
		Embeddings: &mockEmbeddingClient{},
		Files:      fs,
		LLM:        llm,
		Notes:      ns,
		ProgressFn: noOpProgress,
	})

	// Act
	err := svc.Run()

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "processed paths length must be 1", len(fs.processedPaths), 1)
	assert.That(t, "saved notes length must be 0", len(ns.notes), 0)
}

func TestService_Run_MultipleNotesPerFile_SavesAll(t *testing.T) {
	// Arrange
	fs := newMockFileStore()
	fs.files = []extraction.File{
		{Hash: "hash1", Path: "/test/file1.md", Status: extraction.FilePending},
	}
	fs.fileContents["/test/file1.md"] = testFileContent
	llm := &mockLLMClient{
		extractFunc: func(filePath extraction.FilePath, contents string) ([]extraction.MemoryNote, error) {
			return []extraction.MemoryNote{
				{ID: "note-1", Content: "Note 1", Kind: extraction.NoteLearning, Path: filePath},
				{ID: "note-2", Content: "Note 2", Kind: extraction.NotePattern, Path: filePath},
				{ID: "note-3", Content: "Note 3", Kind: extraction.NoteCookbook, Path: filePath},
			}, nil
		},
	}
	ns := &mockNoteStore{}
	svc, _ := extraction.NewService(extraction.ServiceConfig{
		Embeddings: &mockEmbeddingClient{},
		Files:      fs,
		LLM:        llm,
		Notes:      ns,
		ProgressFn: noOpProgress,
	})

	// Act
	err := svc.Run()

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "saved notes length must be 3", len(ns.notes), 3)
}

func TestService_Run_MarkErrorFails_ReturnsError(t *testing.T) {
	// Arrange
	fs := newMockFileStore()
	fs.files = []extraction.File{
		{Hash: "hash1", Path: "/test/missing.md", Status: extraction.FilePending},
	}
	fs.markErrorFunc = func(path extraction.FilePath, reason string) error {
		return errors.New("mark error failed")
	}
	svc, _ := extraction.NewService(extraction.ServiceConfig{
		Embeddings: &mockEmbeddingClient{},
		Files:      fs,
		LLM:        &mockLLMClient{},
		Notes:      &mockNoteStore{},
		ProgressFn: noOpProgress,
	})

	// Act
	err := svc.Run()

	// Assert
	assert.That(t, "err must not be nil when MarkError fails", err != nil, true)
}

func TestService_Run_EmbeddingPreservesNoteData_ReturnsCorrectData(t *testing.T) {
	// Arrange
	fs := newMockFileStore()
	fs.files = []extraction.File{
		{Hash: "hash1", Path: "/test/file1.md", Status: extraction.FilePending},
	}
	fs.fileContents["/test/file1.md"] = testFileContent
	expectedNote := extraction.MemoryNote{
		Content: "Test note content",
		ID:      "test-id",
		Kind:    extraction.NoteDecision,
		Path:    "/test/file1.md",
	}
	llm := &mockLLMClient{
		extractFunc: func(filePath extraction.FilePath, contents string) ([]extraction.MemoryNote, error) {
			return []extraction.MemoryNote{expectedNote}, nil
		},
	}
	var capturedNote extraction.MemoryNote
	ec := &mockEmbeddingClient{
		embedFunc: func(note extraction.MemoryNote) (extraction.EmbeddedNote, error) {
			capturedNote = note
			return extraction.EmbeddedNote{
				Embedding: []float32{0.1, 0.2},
				Note:      note,
			}, nil
		},
	}
	ns := &mockNoteStore{}
	svc, _ := extraction.NewService(extraction.ServiceConfig{
		Embeddings: ec,
		Files:      fs,
		LLM:        llm,
		Notes:      ns,
		ProgressFn: noOpProgress,
	})

	// Act
	err := svc.Run()

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "captured note ID must match", capturedNote.ID, expectedNote.ID)
	assert.That(t, "captured note content must match", capturedNote.Content, expectedNote.Content)
	assert.That(t, "captured note kind must match", capturedNote.Kind, expectedNote.Kind)
}

func TestService_Run_PartialFileFailure_ContinuesProcessing(t *testing.T) {
	// Arrange
	fs := newMockFileStore()
	fs.files = []extraction.File{
		{Hash: "hash1", Path: "/test/missing.md", Status: extraction.FilePending},
		{Hash: "hash2", Path: "/test/valid.md", Status: extraction.FilePending},
	}
	fs.fileContents["/test/valid.md"] = "Valid content"
	ns := &mockNoteStore{}
	svc, _ := extraction.NewService(extraction.ServiceConfig{
		Embeddings: &mockEmbeddingClient{},
		Files:      fs,
		LLM:        &mockLLMClient{},
		Notes:      ns,
		ProgressFn: noOpProgress,
	})

	// Act
	err := svc.Run()

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "error paths length must be 1", len(fs.errorPaths), 1)
	assert.That(t, "saved notes length must be 1 from valid file", len(ns.notes), 1)
}
