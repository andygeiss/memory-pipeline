package outbound

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/andygeiss/memory-pipeline/internal/domain/extraction"
)

// Error definitions for the MarkdownWriter adapter.
var (
	ErrMarkdownWriterEmptyPath = errors.New("outbound: markdown_writer path cannot be empty")
)

// MarkdownWriter is an implementation of the extraction.DocWriter interface.
// It generates human-readable Markdown documentation organized by note kind.
type MarkdownWriter struct {
	notes map[extraction.NoteKind][]extraction.MemoryNote
	path  string
	mu    sync.Mutex
}

// NewMarkdownWriter creates a new instance of MarkdownWriter.
func NewMarkdownWriter(path string) (*MarkdownWriter, error) {
	if path == "" {
		return nil, ErrMarkdownWriterEmptyPath
	}

	return &MarkdownWriter{
		notes: make(map[extraction.NoteKind][]extraction.MemoryNote),
		path:  path,
	}, nil
}

// WriteDoc collects a note for later documentation generation.
func (a *MarkdownWriter) WriteDoc(note extraction.MemoryNote) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.notes[note.Kind] = append(a.notes[note.Kind], note)
	return nil
}

// Finalize writes all collected notes to Markdown files.
func (a *MarkdownWriter) Finalize() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Ensure the docs directory exists.
	if err := os.MkdirAll(a.path, 0750); err != nil {
		return err
	}

	// Write the index file.
	if err := a.writeIndex(); err != nil {
		return err
	}

	// Write each category file.
	categories := []struct {
		kind        extraction.NoteKind
		title       string
		description string
		filename    string
	}{
		{extraction.NoteLearning, "Learnings", "General knowledge, facts, and concepts extracted from the codebase.", "learnings.md"},
		{extraction.NotePattern, "Patterns", "Reusable patterns, best practices, and conventions found in the code.", "patterns.md"},
		{extraction.NoteCookbook, "Cookbooks", "Step-by-step instructions and recipes for common tasks.", "cookbooks.md"},
		{extraction.NoteDecision, "Decisions", "Architectural decisions, trade-offs, and rationale.", "decisions.md"},
	}

	for _, cat := range categories {
		if err := a.writeCategoryFile(cat.kind, cat.title, cat.description, cat.filename); err != nil {
			return err
		}
	}

	return nil
}

// writeIndex creates the main index.md file with links to all categories.
func (a *MarkdownWriter) writeIndex() error {
	var sb strings.Builder

	sb.WriteString("# Knowledge Base\n\n")
	sb.WriteString("This documentation was automatically generated from source code analysis.\n\n")
	sb.WriteString("## Categories\n\n")

	categories := []struct {
		kind     extraction.NoteKind
		title    string
		desc     string
		filename string
	}{
		{extraction.NoteLearning, "Learnings", "General knowledge, facts, and concepts", "learnings.md"},
		{extraction.NotePattern, "Patterns", "Reusable patterns and best practices", "patterns.md"},
		{extraction.NoteCookbook, "Cookbooks", "Step-by-step instructions and recipes", "cookbooks.md"},
		{extraction.NoteDecision, "Decisions", "Architectural decisions and rationale", "decisions.md"},
	}

	for _, cat := range categories {
		count := len(a.notes[cat.kind])
		sb.WriteString(fmt.Sprintf("- [%s](%s) (%d notes) - %s\n", cat.title, cat.filename, count, cat.desc))
	}

	// Write summary statistics.
	totalNotes := 0
	for _, notes := range a.notes {
		totalNotes += len(notes)
	}
	sb.WriteString(fmt.Sprintf("\n## Summary\n\n**Total Notes:** %d\n", totalNotes))

	return os.WriteFile(filepath.Join(a.path, "index.md"), []byte(sb.String()), 0600)
}

// writeCategoryFile writes a single category Markdown file.
func (a *MarkdownWriter) writeCategoryFile(kind extraction.NoteKind, title, description, filename string) error {
	notes := a.notes[kind]

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", title))
	sb.WriteString(description + "\n\n")

	if len(notes) == 0 {
		sb.WriteString("*No notes in this category yet.*\n")
		return os.WriteFile(filepath.Join(a.path, filename), []byte(sb.String()), 0600)
	}

	// Group notes by source file path.
	notesByPath := make(map[extraction.FilePath][]extraction.MemoryNote)
	for _, note := range notes {
		notesByPath[note.Path] = append(notesByPath[note.Path], note)
	}

	// Sort paths for consistent output.
	paths := make([]extraction.FilePath, 0, len(notesByPath))
	for p := range notesByPath {
		paths = append(paths, p)
	}
	slices.Sort(paths)

	// Write notes grouped by file.
	for _, path := range paths {
		pathNotes := notesByPath[path]
		sb.WriteString(fmt.Sprintf("## %s\n\n", path))

		for _, note := range pathNotes {
			sb.WriteString(fmt.Sprintf("%s\n\n", note.Content))
			sb.WriteString("---\n\n")
		}
	}

	return os.WriteFile(filepath.Join(a.path, filename), []byte(sb.String()), 0600)
}
