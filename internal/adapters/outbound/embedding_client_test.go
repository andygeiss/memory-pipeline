package outbound_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/memory-pipeline/internal/adapters/outbound"
	"github.com/andygeiss/memory-pipeline/internal/domain/extraction"
)

const (
	testAPIKey     = "test-api-key"
	testBaseURL    = "http://localhost"
	testEmbedModel = "text-embedding-3-small"
)

func TestEmbeddingClient_New_EmptyAPIKey_ReturnsError(t *testing.T) {
	// Arrange
	apiKey := ""

	// Act
	_, err := outbound.NewEmbeddingClient(apiKey, testBaseURL, testEmbedModel)

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrEmbeddingClientEmptyAPIKey", errors.Is(err, outbound.ErrEmbeddingClientEmptyAPIKey), true)
}

func TestEmbeddingClient_New_EmptyBaseURL_ReturnsError(t *testing.T) {
	// Arrange
	baseURL := ""

	// Act
	_, err := outbound.NewEmbeddingClient(testAPIKey, baseURL, testEmbedModel)

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrEmbeddingClientEmptyBaseURL", errors.Is(err, outbound.ErrEmbeddingClientEmptyBaseURL), true)
}

func TestEmbeddingClient_New_EmptyModel_ReturnsError(t *testing.T) {
	// Arrange
	model := ""

	// Act
	_, err := outbound.NewEmbeddingClient(testAPIKey, testBaseURL, model)

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrEmbeddingClientEmptyModel", errors.Is(err, outbound.ErrEmbeddingClientEmptyModel), true)
}

func TestEmbeddingClient_New_ValidConfig_ReturnsInstance(t *testing.T) {
	// Arrange & Act
	client, err := outbound.NewEmbeddingClient(testAPIKey, testBaseURL, testEmbedModel)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "client must not be nil", client != nil, true)
}

func TestEmbeddingClient_Embed_EmptyContent_ReturnsError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server should not be called for empty content")
	}))
	defer server.Close()
	client, _ := outbound.NewEmbeddingClient(testAPIKey, server.URL, testEmbedModel)
	note := extraction.MemoryNote{
		ID:      "note-1",
		Content: "",
		Kind:    extraction.NoteLearning,
		Path:    "/test/file.md",
	}

	// Act
	_, err := client.Embed(note)

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrEmbeddingClientEmptyText", errors.Is(err, outbound.ErrEmbeddingClientEmptyText), true)
}

func TestEmbeddingClient_Embed_ValidNote_ReturnsEmbeddedNote(t *testing.T) {
	// Arrange
	expectedEmbedding := []float32{0.1, 0.2, 0.3, 0.4}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"data": []map[string]any{
				{"embedding": expectedEmbedding, "index": 0},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	client, _ := outbound.NewEmbeddingClient(testAPIKey, server.URL, testEmbedModel)
	note := extraction.MemoryNote{
		ID:      "note-1",
		Content: "Test content for embedding",
		Kind:    extraction.NoteLearning,
		Path:    "/test/file.md",
	}

	// Act
	result, err := client.Embed(note)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "note ID must match", result.Note.ID, note.ID)
	assert.That(t, "note content must match", result.Note.Content, note.Content)
	assert.That(t, "embedding length must match", len(result.Embedding), len(expectedEmbedding))
}

func TestEmbeddingClient_Embed_ValidNote_SendsCorrectRequest(t *testing.T) {
	// Arrange
	var receivedRequest map[string]any
	var receivedAuthHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthHeader = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&receivedRequest)
		resp := map[string]any{
			"data": []map[string]any{
				{"embedding": []float32{0.1}, "index": 0},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	client, _ := outbound.NewEmbeddingClient(testAPIKey, server.URL, testEmbedModel)
	note := extraction.MemoryNote{
		ID:      "note-1",
		Content: "Test content",
		Kind:    extraction.NoteLearning,
		Path:    "/test/file.md",
	}

	// Act
	_, err := client.Embed(note)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "auth header must be correct", receivedAuthHeader, "Bearer "+testAPIKey)
	assert.That(t, "model must be correct", receivedRequest["model"], testEmbedModel)
	assert.That(t, "input must be correct", receivedRequest["input"], string(note.Content))
}

func TestEmbeddingClient_Embed_ServerError_ReturnsError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal server error"))
	}))
	defer server.Close()
	client, _ := outbound.NewEmbeddingClient(testAPIKey, server.URL, testEmbedModel)
	note := extraction.MemoryNote{
		ID:      "note-1",
		Content: "Test content",
		Kind:    extraction.NoteLearning,
		Path:    "/test/file.md",
	}

	// Act
	_, err := client.Embed(note)

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrEmbeddingClientResponse", errors.Is(err, outbound.ErrEmbeddingClientResponse), true)
}

func TestEmbeddingClient_Embed_APIError_ReturnsError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"error": map[string]any{
				"message": "Invalid API key",
				"code":    "invalid_api_key",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	client, _ := outbound.NewEmbeddingClient(testAPIKey, server.URL, testEmbedModel)
	note := extraction.MemoryNote{
		ID:      "note-1",
		Content: "Test content",
		Kind:    extraction.NoteLearning,
		Path:    "/test/file.md",
	}

	// Act
	_, err := client.Embed(note)

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrEmbeddingClientResponse", errors.Is(err, outbound.ErrEmbeddingClientResponse), true)
}

func TestEmbeddingClient_Embed_EmptyData_ReturnsError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"data": []map[string]any{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	client, _ := outbound.NewEmbeddingClient(testAPIKey, server.URL, testEmbedModel)
	note := extraction.MemoryNote{
		ID:      "note-1",
		Content: "Test content",
		Kind:    extraction.NoteLearning,
		Path:    "/test/file.md",
	}

	// Act
	_, err := client.Embed(note)

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrEmbeddingClientResponse", errors.Is(err, outbound.ErrEmbeddingClientResponse), true)
}

func TestEmbeddingClient_Embed_InvalidJSON_ReturnsError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()
	client, _ := outbound.NewEmbeddingClient(testAPIKey, server.URL, testEmbedModel)
	note := extraction.MemoryNote{
		ID:      "note-1",
		Content: "Test content",
		Kind:    extraction.NoteLearning,
		Path:    "/test/file.md",
	}

	// Act
	_, err := client.Embed(note)

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrEmbeddingClientResponse", errors.Is(err, outbound.ErrEmbeddingClientResponse), true)
}

func TestEmbeddingClient_Embed_UnauthorizedStatus_ReturnsError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer server.Close()
	client, _ := outbound.NewEmbeddingClient("invalid-api-key", server.URL, testEmbedModel)
	note := extraction.MemoryNote{
		ID:      "note-1",
		Content: "Test content",
		Kind:    extraction.NoteLearning,
		Path:    "/test/file.md",
	}

	// Act
	_, err := client.Embed(note)

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrEmbeddingClientResponse", errors.Is(err, outbound.ErrEmbeddingClientResponse), true)
}

func TestEmbeddingClient_Embed_UsesCorrectEndpoint_SendsToEmbeddingsPath(t *testing.T) {
	// Arrange
	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		resp := map[string]any{
			"data": []map[string]any{
				{"embedding": []float32{0.1}, "index": 0},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	client, _ := outbound.NewEmbeddingClient(testAPIKey, server.URL, testEmbedModel)
	note := extraction.MemoryNote{
		ID:      "note-1",
		Content: "Test content",
		Kind:    extraction.NoteLearning,
		Path:    "/test/file.md",
	}

	// Act
	_, err := client.Embed(note)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "path must be /embeddings", receivedPath, "/embeddings")
}

func TestEmbeddingClient_Embed_DifferentNoteKinds_PreservesKind(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"data": []map[string]any{
				{"embedding": []float32{0.1, 0.2}, "index": 0},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	client, _ := outbound.NewEmbeddingClient(testAPIKey, server.URL, testEmbedModel)
	testCases := []extraction.NoteKind{
		extraction.NoteLearning,
		extraction.NotePattern,
		extraction.NoteCookbook,
		extraction.NoteDecision,
	}

	for _, kind := range testCases {
		note := extraction.MemoryNote{
			ID:      "note-1",
			Content: "Test content",
			Kind:    kind,
			Path:    "/test/file.md",
		}

		// Act
		result, err := client.Embed(note)

		// Assert
		assert.That(t, "err must be nil for kind "+string(kind), err, nil)
		assert.That(t, "kind must be preserved for "+string(kind), result.Note.Kind, kind)
	}
}
