package extraction

import "errors"

var (
	ErrServiceConfigMissingEmbeddingClient = errors.New("extraction: service_config is missing embedding client")
	ErrServiceConfigMissingFileStore       = errors.New("extraction: service_config is missing file store")
	ErrServiceConfigMissingLLMClient       = errors.New("extraction: service_config is missing LLM client")
	ErrServiceConfigMissingNoteStore       = errors.New("extraction: service_config is missing note store")
)

// ServiceConfig holds the dependencies required to create a new extraction Service.
type ServiceConfig struct {
	Embeddings EmbeddingClient
	Files      FileStore
	LLM        LLMClient
	Notes      NoteStore
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
	return nil
}

// Service represents the main service for extracting notes from files.
// It orchestrates the process of fetching files, extracting notes using an LLM,
// embedding the notes, and storing them.
type Service struct {
	ec EmbeddingClient
	fs FileStore
	lc LLMClient
	ns NoteStore
}

// NewService creates a new instance of the extraction Service.
func NewService(cfg ServiceConfig) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Service{
		ec: cfg.Embeddings,
		fs: cfg.Files,
		lc: cfg.LLM,
		ns: cfg.Notes,
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
		file, err := a.fs.NextPending()
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
		if err := a.fs.MarkProcessing(file.Path); err != nil {
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

	for _, file := range files {
		// Read file contents.
		contents, err := a.fs.ReadFile(file.Path)
		if err != nil {
			if markErr := a.fs.MarkError(file.Path, err.Error()); markErr != nil {
				return nil, markErr
			}
			continue
		}

		// Extract notes from content.
		notes, err := a.lc.ExtractNotes(file.Path, contents)
		if err != nil {
			if markErr := a.fs.MarkError(file.Path, err.Error()); markErr != nil {
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

	for _, note := range notes {
		embedded, err := a.ec.Embed(note)
		if err != nil {
			return nil, err
		}
		embeddedNotes = append(embeddedNotes, embedded)
	}

	return embeddedNotes, nil
}

// saveNotes persists the embedded notes to the NoteStore.
func (a *Service) saveNotes(notes []EmbeddedNote) error {
	for _, note := range notes {
		if err := a.ns.SaveNote(note); err != nil {
			return err
		}
	}
	return nil
}

// updateFileStatus marks all files as processed.
func (a *Service) updateFileStatus(files []File) error {
	for _, file := range files {
		if err := a.fs.MarkProcessed(file.Path); err != nil {
			return err
		}
	}
	return nil
}
