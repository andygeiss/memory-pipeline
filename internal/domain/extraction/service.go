package extraction

import "errors"

var (
	ErrServiceConfigMissingEmbeddingClient = errors.New("extraction: service_config is missing embedding client")
	ErrServiceConfigMissingFileStore       = errors.New("extraction: service_config is missing file store")
	ErrServiceConfigMissingLLMClient       = errors.New("extraction: service_config is missing LLM client")
	ErrServiceConfigMissingNoteStore       = errors.New("extraction: service_config is missing note store")
	ErrServiceConfigMissingProgressBar     = errors.New("extraction: service_config is missing progress bar")
)

// ProgressFn defines a function type for reporting progress.
type ProgressFn func(current, total int, desc string)

// ServiceConfig holds the dependencies required to create a new extraction Service.
type ServiceConfig struct {
	Embeddings EmbeddingClient
	Files      FileStore
	LLM        LLMClient
	Notes      NoteStore
	ProgressFn ProgressFn
}

// Validate checks if the ServiceConfig has all required dependencies set.
func (a ServiceConfig) Validate() error {
	if a.Embeddings == nil {
		return ErrServiceConfigMissingEmbeddingClient
	}
	if a.Files == nil {
		return ErrServiceConfigMissingFileStore
	}
	if a.LLM == nil {
		return ErrServiceConfigMissingLLMClient
	}
	if a.Notes == nil {
		return ErrServiceConfigMissingNoteStore
	}
	if a.ProgressFn == nil {
		return ErrServiceConfigMissingProgressBar
	}
	return nil
}

// Service represents the main service for extracting notes from files.
// It orchestrates the process of fetching files, extracting notes using an LLM,
// embedding the notes, and storing them.
type Service struct {
	// embeddingClient generates vector embeddings for memory notes.
	embeddingClient EmbeddingClient
	// fileStore manages file discovery, reading, and status tracking.
	fileStore FileStore
	// llmClient extracts structured notes from file contents.
	llmClient LLMClient
	// noteStore persists embedded notes to storage.
	noteStore NoteStore
	// progressFn reports progress updates during pipeline execution.
	progressFn ProgressFn
}

// NewService creates a new instance of the extraction Service.
func NewService(cfg ServiceConfig) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Service{
		embeddingClient: cfg.Embeddings,
		fileStore:       cfg.Files,
		llmClient:       cfg.LLM,
		noteStore:       cfg.Notes,
		progressFn:      cfg.ProgressFn,
	}, nil
}

// Run starts the extraction service to process files and extract notes.
// It uses a sequential pipeline pattern for processing:
// 1. Fetch pending files from the FileStore.
// 2. For each file, read its content and extract notes using the LLMClient.
// 3. Embed the notes using the EmbeddingClient.
// 4. Store the notes in the NoteStore.
// 5. Update the file status in the FileStore.
func (a *Service) Run() error {
	// 1. Fetch pending files from the FileStore.
	files, err := a.collectPendingFiles()
	if err != nil {
		return err
	}

	// If there are no files to process, return early.
	if len(files) == 0 {
		return nil
	}

	// 2. For each file, read its content and extract notes using the LLMClient.
	notes, err := a.extractNotes(files)
	if err != nil {
		return err
	}

	// If no notes were extracted, mark files as processed and return.
	if len(notes) == 0 {
		return a.updateFileStatus(files)
	}

	// 3. Embed the notes using the EmbeddingClient.
	embeddedNotes, err := a.embedNotes(notes)
	if err != nil {
		return err
	}

	// 4. Store the embedded notes in the NoteStore.
	if err := a.saveNotes(embeddedNotes); err != nil {
		return err
	}

	// 5. Update the file status in the FileStore.
	return a.updateFileStatus(files)
}

// collectPendingFiles retrieves all pending files from the FileStore.
func (a *Service) collectPendingFiles() ([]File, error) {
	var files []File

	for {
		file, err := a.fileStore.NextPending()
		if err != nil {
			// Check for sentinel error indicating no more files.
			if isNoMoreFilesError(err) {
				break
			}
			return nil, err
		}

		// No more pending files (nil file also signals completion).
		if file == nil {
			break
		}

		// Mark file as processing.
		if err := a.fileStore.MarkProcessing(file.Path); err != nil {
			return nil, err
		}

		files = append(files, *file)
	}

	return files, nil
}

// isNoMoreFilesError checks if the error indicates no more pending files.
func isNoMoreFilesError(err error) bool {
	return err != nil && err.Error() == ErrFileStoreNoMoreFiles.Error()
}

// extractNotes reads file contents and extracts notes using the LLM.
func (a *Service) extractNotes(files []File) ([]MemoryNote, error) {
	var allNotes []MemoryNote
	total := len(files)

	for i, file := range files {
		a.progressFn(i+1, total, "1. Extracting notes")
		// Read file contents.
		contents, err := a.fileStore.ReadFile(file.Path)
		if err != nil {
			if markErr := a.fileStore.MarkError(file.Path, err.Error()); markErr != nil {
				return nil, markErr
			}
			continue
		}

		// Extract notes from content.
		notes, err := a.llmClient.ExtractNotes(file.Path, contents)
		if err != nil {
			if markErr := a.fileStore.MarkError(file.Path, err.Error()); markErr != nil {
				return nil, markErr
			}
			continue
		}

		allNotes = append(allNotes, notes...)
	}

	return allNotes, nil
}

// embedNotes generates embeddings for each note.
func (a *Service) embedNotes(notes []MemoryNote) ([]EmbeddedNote, error) {
	embeddedNotes := make([]EmbeddedNote, 0, len(notes))
	total := len(notes)

	for i, note := range notes {
		a.progressFn(i+1, total, "2. Embedding notes")
		embedded, err := a.embeddingClient.Embed(note)
		if err != nil {
			return nil, err
		}
		embeddedNotes = append(embeddedNotes, embedded)
	}

	return embeddedNotes, nil
}

// saveNotes persists the embedded notes to the NoteStore.
func (a *Service) saveNotes(notes []EmbeddedNote) error {
	total := len(notes)

	for i, note := range notes {
		a.progressFn(i+1, total, "3. Saving notes")
		if err := a.noteStore.SaveNote(note); err != nil {
			return err
		}
	}
	return nil
}

// updateFileStatus marks all files as processed.
func (a *Service) updateFileStatus(files []File) error {
	total := len(files)

	for i, file := range files {
		a.progressFn(i+1, total, "4. Updating status")
		if err := a.fileStore.MarkProcessed(file.Path); err != nil {
			return err
		}
	}
	return nil
}
