package extraction

import "errors"

// ErrFileStoreNoMoreFiles is returned when the file store has no more pending files.
var ErrFileStoreNoMoreFiles = errors.New("extraction: file_store has no more pending files")

// We group the definitions related to files and notes here for better organization.
// This helps in maintaining clarity and separation of concerns within the extraction domain.
// It also facilitates easier imports and usage across different parts of the application.

// === File Definitions ===

// FileHash represents a unique hash identifier for a file.
type FileHash string

// FilePath represents the path to a file.
type FilePath string

// FileStatus represents the processing status of a file.
type FileStatus string

const (
	// FilePending indicates the file is waiting to be processed.
	FilePending FileStatus = "pending"
	// FileProcessing indicates the file is currently being processed.
	FileProcessing FileStatus = "processing"
	// FileProcessed indicates the file has been successfully processed.
	FileProcessed FileStatus = "processed"
	// FileError indicates the file encountered an error during processing.
	FileError FileStatus = "error"
)

// File represents a file with its hash, path, and processing status.
type File struct {
	Hash   FileHash
	Path   FilePath
	Status FileStatus
}

// === Note Definitions ===

// NoteContent represents the textual content of a note.
type NoteContent string

// NodeID represents a unique identifier for a node in the knowledge graph.
type NodeID string

// NoteKind represents the category or type of a note.
type NoteKind string

const (
	// NoteLearning represents general knowledge, facts, or concepts.
	NoteLearning NoteKind = "learning"
	// NotePattern represents reusable patterns, best practices, or conventions.
	NotePattern NoteKind = "pattern"
	// NoteCookbook represents step-by-step instructions or recipes.
	NoteCookbook NoteKind = "cookbook"
	// NoteDecision represents architectural decisions, trade-offs, or rationale.
	NoteDecision NoteKind = "decision"
)

// MemoryNote represents a note stored in memory with its metadata.
type MemoryNote struct {
	ID      NodeID
	Content NoteContent
	Kind    NoteKind
	Path    FilePath
}

// EmbeddedNote represents a note in the knowledge graph with its embedding vector.
type EmbeddedNote struct {
	Note      MemoryNote
	Embedding []float32
}
