package extraction

// EmbeddingClient defines the interface for generating embeddings from notes.
type EmbeddingClient interface {
	Embed(note MemoryNote) (EmbeddedNote, error)
}

// FileStore defines the interface for storing and managing files.
type FileStore interface {
	MarkError(path FilePath, reason string) error
	NextPending() (*File, error)
	MarkProcessed(path FilePath) error
	MarkProcessing(path FilePath) error
	ReadFile(path FilePath) (string, error)
}

// LLMClient defines the interface for interacting with a large language model to extract notes.
type LLMClient interface {
	ExtractNotes(filePath FilePath, contents string) ([]MemoryNote, error)
}

// NoteStore defines the interface for storing embedded notes.
type NoteStore interface {
	SaveNote(note EmbeddedNote) error
}
