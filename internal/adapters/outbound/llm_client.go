package outbound

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/andygeiss/memory-pipeline/internal/domain/extraction"
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

	// Request extraction from the LLM.
	extracted, err := a.requestExtraction(contents)
	if err != nil {
		return nil, err
	}

	// Convert extracted notes to MemoryNote type.
	// This maps the extracted notes to the domain model.
	notes := make([]extraction.MemoryNote, len(extracted.Notes))
	for i, note := range extracted.Notes {
		notes[i] = extraction.MemoryNote{
			Content: extraction.NoteContent(note.Content),
			ID:      extraction.NodeID(note.ID),
			Kind:    parseNoteKind(note.Kind),
			Path:    filePath,
		}
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
const systemPrompt = `You are a senior staff-level knowledge extraction assistant helping developers build a long-term project memory.

Your task:
Analyze the provided content and extract only high-value, reusable knowledge as structured memory notes.

For each distinct piece of knowledge, create a note with:
- id: Leave this as an empty string "" OR a short descriptive slug. A stable unique ID (note-<uuid>) will be added later by the system.
- kind: One of "learning", "pattern", "cookbook", or "decision"
- content: A clear, self-contained description of the knowledge that makes sense without seeing the original file

Note kinds (clear, typed schema):
- learning: General knowledge, facts, or concepts that explain what something is or why it matters.
- pattern: Reusable patterns, best practices, or conventions that a developer could apply in other places.
- cookbook: Step-by-step instructions or recipes describing how to do something, in ordered steps.
- decision: Architectural decisions, trade-offs, or rationale, ideally including context, options considered, and the chosen direction.

Note quality over volume:
- Prefer fewer, higher-quality notes over many trivial ones.
- Only create a note if it would still be useful to a developer weeks later when reading it out of context.
- Each note should capture exactly one main idea; split unrelated ideas into separate notes.
- Avoid restating code or comments line-by-line; capture the underlying intent, principle, or decision instead.
- Make every note self-contained: avoid phrases like "in this file" or "above code"; write it so it stands on its own.
- Do not invent details that are not clearly supported by the content.

Few-shot style examples (follow these styles, not their content):

Example "learning":
{
  "id": "",
  "kind": "learning",
  "content": "The service uses structured logging with consistent log levels to make production issues easier to filter and diagnose."
}

Example "pattern":
{
  "id": "note-hexagonal-ports-adapters",
  "kind": "pattern",
  "content": "The codebase follows hexagonal architecture by defining ports in the domain layer and implementing adapters at the boundaries for external systems."
}

Example "cookbook":
{
  "id": "",
  "kind": "cookbook",
  "content": "To run the pipeline locally: (1) start the local LLM server, (2) configure environment variables, (3) run the CLI command, and (4) inspect the generated state and notes files."
}

Example "decision":
{
  "id": "note-use-local-llm",
  "kind": "decision",
  "content": "The team chose a local OpenAI-compatible LLM instead of a remote API to reduce latency, avoid external dependencies, and keep code private."
}

Extraction principles:
- Use the definitions and examples above to choose the most appropriate kind for each note.
- If a piece of knowledge could fit multiple kinds, choose the one that is most helpful for future reuse (pattern or decision is often better than learning).
- Keep wording concise but clear; optimize for future retrieval and understanding.
- The system will generate final unique IDs; you do not need to create UUIDs.

Formatting rules (strict):
- Respond with a single JSON object containing a "notes" array.
- Each element in "notes" must have exactly these fields: "id", "kind", "content".
- Do not include any other top-level keys or fields.
- Do not include explanations, commentary, or markdown.
- Do not wrap the JSON in code fences.

Example response:
{
  "notes": [
    {
      "id": "",
      "kind": "learning",
      "content": "..."
    },
    {
      "id": "note-logging-strategy",
      "kind": "pattern",
      "content": "..."
    }
  ]
}

If no meaningful, reusable notes can be extracted, respond exactly with:
{"notes": []}`
