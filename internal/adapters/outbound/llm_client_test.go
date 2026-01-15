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
	testLLMAuth     = "test-llm-auth-value"
	testLLMBaseURL  = "http://localhost"
	testLLMModel    = "gpt-4"
	testLLMFilePath = "/test/file.md"
)

func TestLLMClient_New_EmptyAPIKey_ReturnsError(t *testing.T) {
	// Arrange
	apiKey := ""

	// Act
	_, err := outbound.NewLLMClient(apiKey, testLLMBaseURL, testLLMModel)

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrLLMClientEmptyAPIKey", errors.Is(err, outbound.ErrLLMClientEmptyAPIKey), true)
}

func TestLLMClient_New_EmptyBaseURL_ReturnsError(t *testing.T) {
	// Arrange
	baseURL := ""

	// Act
	_, err := outbound.NewLLMClient(testLLMAuth, baseURL, testLLMModel)

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrLLMClientEmptyBaseURL", errors.Is(err, outbound.ErrLLMClientEmptyBaseURL), true)
}

func TestLLMClient_New_EmptyModel_ReturnsError(t *testing.T) {
	// Arrange
	model := ""

	// Act
	_, err := outbound.NewLLMClient(testLLMAuth, testLLMBaseURL, model)

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrLLMClientEmptyModel", errors.Is(err, outbound.ErrLLMClientEmptyModel), true)
}

func TestLLMClient_New_ValidConfig_ReturnsInstance(t *testing.T) {
	// Arrange & Act
	client, err := outbound.NewLLMClient(testLLMAuth, testLLMBaseURL, testLLMModel)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "client must not be nil", client != nil, true)
}

func TestLLMClient_ExtractNotes_EmptyContents_ReturnsError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server should not be called for empty contents")
	}))
	defer server.Close()
	client, _ := outbound.NewLLMClient(testLLMAuth, server.URL, testLLMModel)

	// Act
	_, err := client.ExtractNotes(testLLMFilePath, "")

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrLLMClientEmptyContents", errors.Is(err, outbound.ErrLLMClientEmptyContents), true)
}

func TestLLMClient_ExtractNotes_ValidContents_ReturnsNotes(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": `{"notes": [{"id": "note-1", "kind": "learning", "content": "Test learning note"}]}`,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	client, _ := outbound.NewLLMClient(testLLMAuth, server.URL, testLLMModel)

	// Act
	notes, err := client.ExtractNotes(testLLMFilePath, "Some test content")

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "notes length must be 1", len(notes), 1)
	assert.That(t, "note ID must not be empty", notes[0].ID != "", true)
	assert.That(t, "note kind must be NoteLearning", notes[0].Kind, extraction.NoteLearning)
	assert.That(t, "note content must match", notes[0].Content, extraction.NoteContent("Test learning note"))
	assert.That(t, "note path must match", notes[0].Path, extraction.FilePath(testLLMFilePath))
}

func TestLLMClient_ExtractNotes_MultipleNotes_ReturnsAll(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notesJSON := `{"notes": [` +
			`{"id": "note-1", "kind": "learning", "content": "Learning note"},` +
			`{"id": "note-2", "kind": "pattern", "content": "Pattern note"},` +
			`{"id": "note-3", "kind": "cookbook", "content": "Cookbook note"},` +
			`{"id": "note-4", "kind": "decision", "content": "Decision note"}` +
			`]}`
		resp := map[string]any{
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": notesJSON,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	client, _ := outbound.NewLLMClient(testLLMAuth, server.URL, testLLMModel)

	// Act
	notes, err := client.ExtractNotes(testLLMFilePath, "Some test content")

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "notes length must be 4", len(notes), 4)
	expectedKinds := []extraction.NoteKind{
		extraction.NoteLearning,
		extraction.NotePattern,
		extraction.NoteCookbook,
		extraction.NoteDecision,
	}
	for i, kind := range expectedKinds {
		assert.That(t, "kind at index must match", notes[i].Kind, kind)
	}
}

func TestLLMClient_ExtractNotes_EmptyNotes_ReturnsEmptySlice(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": `{"notes": []}`,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	client, _ := outbound.NewLLMClient(testLLMAuth, server.URL, testLLMModel)

	// Act
	notes, err := client.ExtractNotes(testLLMFilePath, "Some test content")

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "notes length must be 0", len(notes), 0)
}

func TestLLMClient_ExtractNotes_ValidRequest_SendsCorrectHeaders(t *testing.T) {
	// Arrange
	var receivedAuthHeader string
	var receivedContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthHeader = r.Header.Get("Authorization")
		receivedContentType = r.Header.Get("Content-Type")
		resp := map[string]any{
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": `{"notes": []}`,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	client, _ := outbound.NewLLMClient(testLLMAuth, server.URL, testLLMModel)

	// Act
	_, err := client.ExtractNotes(testLLMFilePath, "Some test content")

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "auth header must be correct", receivedAuthHeader, "Bearer "+testLLMAuth)
	assert.That(t, "content type must be application/json", receivedContentType, "application/json")
}

func TestLLMClient_ExtractNotes_ValidRequest_SendsCorrectBody(t *testing.T) {
	// Arrange
	var receivedRequest map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&receivedRequest)
		resp := map[string]any{
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": `{"notes": []}`,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	client, _ := outbound.NewLLMClient(testLLMAuth, server.URL, testLLMModel)

	// Act
	_, err := client.ExtractNotes(testLLMFilePath, "Test content to extract")

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "model must be correct", receivedRequest["model"], testLLMModel)
	messages, ok := receivedRequest["messages"].([]any)
	assert.That(t, "messages must be array", ok, true)
	assert.That(t, "messages length must be 2", len(messages), 2)
	userMessage := messages[1].(map[string]any)
	assert.That(t, "user content must be correct", userMessage["content"], "Test content to extract")
}

func TestLLMClient_ExtractNotes_ServerError_ReturnsError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal server error"))
	}))
	defer server.Close()
	client, _ := outbound.NewLLMClient(testLLMAuth, server.URL, testLLMModel)

	// Act
	_, err := client.ExtractNotes(testLLMFilePath, "Some test content")

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrLLMClientResponse", errors.Is(err, outbound.ErrLLMClientResponse), true)
}

func TestLLMClient_ExtractNotes_APIError_ReturnsError(t *testing.T) {
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
	client, _ := outbound.NewLLMClient(testLLMAuth, server.URL, testLLMModel)

	// Act
	_, err := client.ExtractNotes(testLLMFilePath, "Some test content")

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrLLMClientResponse", errors.Is(err, outbound.ErrLLMClientResponse), true)
}

func TestLLMClient_ExtractNotes_EmptyChoices_ReturnsError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"choices": []map[string]any{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	client, _ := outbound.NewLLMClient(testLLMAuth, server.URL, testLLMModel)

	// Act
	_, err := client.ExtractNotes(testLLMFilePath, "Some test content")

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrLLMClientResponse", errors.Is(err, outbound.ErrLLMClientResponse), true)
}

func TestLLMClient_ExtractNotes_InvalidJSON_ReturnsError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()
	client, _ := outbound.NewLLMClient(testLLMAuth, server.URL, testLLMModel)

	// Act
	_, err := client.ExtractNotes(testLLMFilePath, "Some test content")

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrLLMClientResponse", errors.Is(err, outbound.ErrLLMClientResponse), true)
}

func TestLLMClient_ExtractNotes_InvalidNotesJSON_ReturnsError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": `not valid json`,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	client, _ := outbound.NewLLMClient(testLLMAuth, server.URL, testLLMModel)

	// Act
	_, err := client.ExtractNotes(testLLMFilePath, "Some test content")

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrLLMClientResponse", errors.Is(err, outbound.ErrLLMClientResponse), true)
}

func TestLLMClient_ExtractNotes_UsesCorrectEndpoint_SendsToChatCompletions(t *testing.T) {
	// Arrange
	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		resp := map[string]any{
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": `{"notes": []}`,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	client, _ := outbound.NewLLMClient(testLLMAuth, server.URL, testLLMModel)

	// Act
	_, err := client.ExtractNotes(testLLMFilePath, "Some test content")

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "path must be /chat/completions", receivedPath, "/chat/completions")
}

func TestLLMClient_ExtractNotes_UnknownKind_DefaultsToLearning(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notesJSON := `{"notes": [{"id": "note-1", "kind": "unknown_kind", "content": "Test note"}]}`
		resp := map[string]any{
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": notesJSON,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	client, _ := outbound.NewLLMClient(testLLMAuth, server.URL, testLLMModel)

	// Act
	notes, err := client.ExtractNotes(testLLMFilePath, "Some test content")

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "notes length must be 1", len(notes), 1)
	assert.That(t, "unknown kind must default to NoteLearning", notes[0].Kind, extraction.NoteLearning)
}

func TestLLMClient_ExtractNotes_UnauthorizedStatus_ReturnsError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer server.Close()
	client, _ := outbound.NewLLMClient("invalid-api-key", server.URL, testLLMModel)

	// Act
	_, err := client.ExtractNotes(testLLMFilePath, "Some test content")

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
	assert.That(t, "err must be ErrLLMClientResponse", errors.Is(err, outbound.ErrLLMClientResponse), true)
}
