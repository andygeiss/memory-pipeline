package outbound

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/loopforge-ai/memory-pipeline/internal/domain/extraction"
)

// Error definitions for the LLMClient adapter.
var (
	ErrLLMClientEmptyAPIKey   = errors.New("outbound: llm_client api_key cannot be empty")
	ErrLLMClientEmptyBaseURL  = errors.New("outbound: llm_client base_url cannot be empty")
	ErrLLMClientEmptyContents = errors.New("outbound: llm_client contents cannot be empty")
	ErrLLMClientEmptyModel    = errors.New("outbound: llm_client model cannot be empty")
	ErrLLMClientRequest       = errors.New("outbound: llm_client request failed")
	ErrLLMClientResponse      = errors.New("outbound: llm_client response error")
)

// chatRequest represents the request payload for the chat completions API.
type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

// chatMessage represents a single message in the chat.
type chatMessage struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

// chatResponse represents the response from the chat completions API.
type chatResponse struct {
	Error   *llmAPIError `json:"error,omitempty"`
	Choices []chatChoice `json:"choices"`
}

// chatChoice represents a single choice in the chat response.
type chatChoice struct {
	Message chatMessage `json:"message"`
	Index   int         `json:"index"`
}

// llmAPIError represents an error response from the LLM API.
type llmAPIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// extractedNotes represents the JSON structure for extracted notes.
type extractedNotes struct {
	Notes []extractedNote `json:"notes"`
}

// extractedNote represents a single extracted note from the LLM response.
type extractedNote struct {
	Content string `json:"content"`
	ID      string `json:"id"`
	Kind    string `json:"kind"`
}

// LLMClient is an implementation of a client for interacting with a large language model (LLM).
type LLMClient struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	chatModel  string
}

// NewLLMClient creates a new instance of LLMClient.
func NewLLMClient(apiKey, baseURL, chatModel string) (*LLMClient, error) {
	if apiKey == "" {
		return nil, ErrLLMClientEmptyAPIKey
	}
	if baseURL == "" {
		return nil, ErrLLMClientEmptyBaseURL
	}
	if chatModel == "" {
		return nil, ErrLLMClientEmptyModel
	}

	return &LLMClient{
		httpClient: &http.Client{Timeout: 60 * time.Second},
		apiKey:     apiKey,
		baseURL:    baseURL,
		chatModel:  chatModel,
	}, nil
}

// ExtractNotes uses the LLM to extract memory notes from the given file contents.
func (a *LLMClient) ExtractNotes(filePath extraction.FilePath, contents string) ([]extraction.MemoryNote, error) {
	if contents == "" {
		return nil, ErrLLMClientEmptyContents
	}

	extracted, err := a.requestExtraction(contents)
	if err != nil {
		return nil, err
	}

	notes := make([]extraction.MemoryNote, 0, len(extracted.Notes))
	for _, n := range extracted.Notes {
		notes = append(notes, extraction.MemoryNote{
			Content: extraction.NoteContent(n.Content),
			ID:      extraction.NodeID(n.ID),
			Kind:    parseNoteKind(n.Kind),
			Path:    filePath,
		})
	}

	return notes, nil
}

// requestExtraction sends a request to the chat completions API and returns extracted notes.
func (a *LLMClient) requestExtraction(contents string) (*extractedNotes, error) {
	body, err := a.sendChatRequest(contents)
	if err != nil {
		return nil, err
	}

	return a.parseChatResponse(body)
}

// sendChatRequest sends the chat completion request and returns the response body.
func (a *LLMClient) sendChatRequest(contents string) ([]byte, error) {
	reqBody := chatRequest{
		Messages: []chatMessage{
			{Content: systemPrompt, Role: "system"},
			{Content: contents, Role: "user"},
		},
		Model: a.chatModel,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLLMClientRequest, err)
	}

	req, err := http.NewRequest(http.MethodPost, a.baseURL+"/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLLMClientRequest, err)
	}

	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLLMClientRequest, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLLMClientResponse, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d: %s", ErrLLMClientResponse, resp.StatusCode, string(body))
	}

	return body, nil
}

// parseChatResponse parses the chat response and extracts the notes.
func (a *LLMClient) parseChatResponse(body []byte) (*extractedNotes, error) {
	var chatResp chatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLLMClientResponse, err)
	}

	if chatResp.Error != nil {
		return nil, fmt.Errorf("%w: %s", ErrLLMClientResponse, chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("%w: no choices returned", ErrLLMClientResponse)
	}

	var extracted extractedNotes
	if err := json.Unmarshal([]byte(chatResp.Choices[0].Message.Content), &extracted); err != nil {
		return nil, fmt.Errorf("%w: failed to parse notes: %w", ErrLLMClientResponse, err)
	}

	return &extracted, nil
}

// parseNoteKind converts a string to NoteKind, defaulting to NoteLearning.
func parseNoteKind(kind string) extraction.NoteKind {
	switch kind {
	case string(extraction.NoteCookbook):
		return extraction.NoteCookbook
	case string(extraction.NoteDecision):
		return extraction.NoteDecision
	case string(extraction.NoteLearning):
		return extraction.NoteLearning
	case string(extraction.NotePattern):
		return extraction.NotePattern
	default:
		return extraction.NoteLearning
	}
}

// systemPrompt defines the instruction for the LLM to extract notes.
const systemPrompt = `You are a knowledge extraction assistant. Analyze the provided content and extract structured memory notes.

For each distinct piece of knowledge, create a note with:
- id: A unique identifier (use format: note-<uuid> or descriptive slug)
- kind: One of "learning", "pattern", "cookbook", or "decision"
- content: The extracted knowledge in clear, concise form

Note kinds:
- learning: General knowledge, facts, or concepts
- pattern: Reusable patterns, best practices, or conventions
- cookbook: Step-by-step instructions or recipes
- decision: Architectural decisions, trade-offs, or rationale

Respond with a JSON object containing a "notes" array. Example:
{
  "notes": [
    {"id": "note-1", "kind": "learning", "content": "..."},
    {"id": "note-2", "kind": "pattern", "content": "..."}
  ]
}

If no meaningful notes can be extracted, return {"notes": []}.`
