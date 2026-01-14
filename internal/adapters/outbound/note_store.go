package outbound

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/andygeiss/memory-pipeline/internal/domain/extraction"
)

// Error definitions for the NoteStore adapter.
var (
	ErrNoteStoreEmptyPath = errors.New("outbound: note_store path cannot be empty")
)

// storedNote represents a note persisted to disk.
type storedNote struct {
	Content   extraction.NoteContent `json:"content"`
	ID        extraction.NodeID      `json:"id"`
	Kind      extraction.NoteKind    `json:"kind"`
	Path      extraction.FilePath    `json:"path"`
	Embedding []float32              `json:"embedding"`
}

// NoteStore is an implementation of the extraction.NoteStore interface.
// It persists embedded notes to a JSON file.
type NoteStore struct {
	notes map[extraction.NodeID]*storedNote
	path  string
	mu    sync.RWMutex
}

// NewNoteStore creates a new instance of NoteStore.
func NewNoteStore(path string) (*NoteStore, error) {
	if path == "" {
		return nil, ErrNoteStoreEmptyPath
	}

	ns := &NoteStore{
		notes: make(map[extraction.NodeID]*storedNote),
		path:  path,
	}

	// Load existing notes from file if it exists.
	if err := ns.loadNotes(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	return ns, nil
}

// SaveNote saves the given embedded note.
func (a *NoteStore) SaveNote(note extraction.EmbeddedNote) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.notes[note.Note.ID] = &storedNote{
		Content:   note.Note.Content,
		Embedding: note.Embedding,
		ID:        note.Note.ID,
		Kind:      note.Note.Kind,
		Path:      note.Note.Path,
	}

	return a.saveNotes()
}

// loadNotes loads the notes from the storage file.
func (a *NoteStore) loadNotes() error {
	data, err := os.ReadFile(a.path)
	if err != nil {
		return err
	}

	var notes []*storedNote
	if err := json.Unmarshal(data, &notes); err != nil {
		return err
	}

	for _, n := range notes {
		a.notes[n.ID] = n
	}

	return nil
}

// saveNotes persists the notes to the storage file.
func (a *NoteStore) saveNotes() error {
	notes := make([]*storedNote, 0, len(a.notes))
	for _, n := range a.notes {
		notes = append(notes, n)
	}

	data, err := json.MarshalIndent(notes, "", "  ")
	if err != nil {
		return err
	}

	// Ensure the directory exists.
	dir := filepath.Dir(a.path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	return os.WriteFile(a.path, data, 0600)
}
