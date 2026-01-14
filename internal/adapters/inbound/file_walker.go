package inbound

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/andygeiss/cloud-native-utils/security"
	"github.com/andygeiss/memory-pipeline/internal/domain/extraction"
)

// Error definitions for the FileWalker adapter.
var (
	ErrFileWalkerEmptyExtensions = errors.New("inbound: file_walker extensions cannot be empty")
	ErrFileWalkerEmptySourceDir  = errors.New("inbound: file_walker source_dir cannot be empty")
	ErrFileWalkerEmptyStateFile  = errors.New("inbound: file_walker state_file cannot be empty")
	ErrFileWalkerFileNotFound    = errors.New("inbound: file_walker file not found")
)

// fileState represents the persisted state of a tracked file.
type fileState struct {
	Hash    extraction.FileHash   `json:"hash"`
	Path    extraction.FilePath   `json:"path"`
	Reason  string                `json:"reason,omitempty"`
	Status  extraction.FileStatus `json:"status"`
	ModTime int64                 `json:"mod_time"`
}

// FileWalker is an implementation of FileStore that walks the filesystem.
// It scans for files with specified extensions and tracks their processing state.
type FileWalker struct {
	state      map[extraction.FilePath]*fileState
	sourceDir  string
	stateFile  extraction.FilePath
	extensions []string
	mu         sync.RWMutex
}

// NewFileWalker creates a new instance of FileWalker with the given configuration.
func NewFileWalker(sourceDir string, stateFile extraction.FilePath, extensions []string) (*FileWalker, error) {
	if sourceDir == "" {
		return nil, ErrFileWalkerEmptySourceDir
	}
	if stateFile == "" {
		return nil, ErrFileWalkerEmptyStateFile
	}
	if len(extensions) == 0 {
		return nil, ErrFileWalkerEmptyExtensions
	}

	fw := &FileWalker{
		extensions: extensions,
		sourceDir:  sourceDir,
		state:      make(map[extraction.FilePath]*fileState),
		stateFile:  stateFile,
	}

	// Load existing state from file if it exists.
	if err := fw.loadState(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	return fw, nil
}

// MarkError marks the given file as having encountered an error with a reason.
func (a *FileWalker) MarkError(path extraction.FilePath, reason string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	st, ok := a.state[path]
	if !ok {
		return ErrFileWalkerFileNotFound
	}

	st.Status = extraction.FileError
	st.Reason = reason

	return a.saveState()
}

// MarkProcessed marks the given file as processed.
func (a *FileWalker) MarkProcessed(path extraction.FilePath) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	st, ok := a.state[path]
	if !ok {
		return ErrFileWalkerFileNotFound
	}

	st.Status = extraction.FileProcessed
	st.Reason = ""

	return a.saveState()
}

// MarkProcessing marks the given file as currently being processed.
func (a *FileWalker) MarkProcessing(path extraction.FilePath) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	st, ok := a.state[path]
	if !ok {
		return ErrFileWalkerFileNotFound
	}

	st.Status = extraction.FileProcessing
	st.Reason = ""

	return a.saveState()
}

// NextPending returns the next file that is pending processing.
// It scans the source directory for files with matching extensions,
// updates the internal state, and returns the first pending file.
func (a *FileWalker) NextPending() (*extraction.File, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Scan directory and update state.
	if err := a.scanDirectory(); err != nil {
		return nil, err
	}

	// Find first pending file.
	for _, st := range a.state {
		if st.Status == extraction.FilePending {
			return &extraction.File{
				Hash:   st.Hash,
				Path:   st.Path,
				Status: st.Status,
			}, nil
		}
	}

	return nil, extraction.ErrFileStoreNoMoreFiles
}

// ReadFile reads the content of the file at the given path.
func (a *FileWalker) ReadFile(path extraction.FilePath) (string, error) {
	data, err := os.ReadFile(string(path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrFileWalkerFileNotFound
		}
		return "", err
	}
	return string(data), nil
}

// computeHash computes a hash of the file content using vendor security package.
func (a *FileWalker) computeHash(path string) (extraction.FileHash, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: Path comes from trusted directory walk
	if err != nil {
		return "", err
	}

	hash := security.Hash("file-walker", data)
	return extraction.FileHash(hex.EncodeToString(hash)), nil
}

// hasValidExtension checks if the file has one of the configured extensions.
func (a *FileWalker) hasValidExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return slices.Contains(a.extensions, ext)
}

// loadState loads the processing state from the state file.
func (a *FileWalker) loadState() error {
	data, err := os.ReadFile(string(a.stateFile))
	if err != nil {
		return err
	}

	var states []*fileState
	if err := json.Unmarshal(data, &states); err != nil {
		return err
	}

	for _, st := range states {
		a.state[st.Path] = st
	}

	return nil
}

// saveState persists the processing state to the state file.
func (a *FileWalker) saveState() error {
	states := make([]*fileState, 0, len(a.state))
	for _, st := range a.state {
		states = append(states, st)
	}

	data, err := json.MarshalIndent(states, "", "  ")
	if err != nil {
		return err
	}

	// Ensure the directory exists.
	dir := filepath.Dir(string(a.stateFile))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(string(a.stateFile), data, 0600)
}

// scanDirectory walks the source directory and updates the internal state
// for files with valid extensions.
func (a *FileWalker) scanDirectory() error {
	return filepath.WalkDir(a.sourceDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		// Skip directories and files without valid extensions.
		if d.IsDir() || !a.hasValidExtension(path) {
			return nil
		}

		return a.processDiscoveredFile(path, d)
	})
}

// processDiscoveredFile handles a single file discovered during directory scan.
func (a *FileWalker) processDiscoveredFile(path string, d fs.DirEntry) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Get file info for ModTime.
	info, err := d.Info()
	if err != nil {
		return err
	}
	modTime := info.ModTime().UnixNano()

	filePath := extraction.FilePath(absPath)

	// Check if file is already tracked.
	existing, ok := a.state[filePath]
	if ok {
		return a.updateExistingFile(existing, absPath, modTime)
	}

	// New file: compute hash and add as pending.
	hash, err := a.computeHash(absPath)
	if err != nil {
		return err
	}

	a.state[filePath] = &fileState{
		Hash:    hash,
		Path:    filePath,
		Status:  extraction.FilePending,
		ModTime: modTime,
	}

	return nil
}

// updateExistingFile updates an already tracked file if its content has changed.
func (a *FileWalker) updateExistingFile(existing *fileState, absPath string, modTime int64) error {
	// If ModTime unchanged, skip expensive hash computation.
	if existing.ModTime == modTime {
		return nil
	}

	// ModTime changed, compute hash to verify content change.
	hash, err := a.computeHash(absPath)
	if err != nil {
		return err
	}

	// Update ModTime regardless of hash change.
	existing.ModTime = modTime

	// If hash changed, mark as pending for reprocessing.
	if existing.Hash != hash {
		existing.Hash = hash
		existing.Status = extraction.FilePending
		existing.Reason = ""
	}

	return nil
}
