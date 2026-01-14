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

// Error definitions for the EmbeddingClient adapter.
var (
	ErrEmbeddingClientEmptyAPIKey  = errors.New("outbound: embedding_client api_key cannot be empty")
	ErrEmbeddingClientEmptyBaseURL = errors.New("outbound: embedding_client base_url cannot be empty")
	ErrEmbeddingClientEmptyModel   = errors.New("outbound: embedding_client model cannot be empty")
	ErrEmbeddingClientEmptyText    = errors.New("outbound: embedding_client text cannot be empty")
	ErrEmbeddingClientRequest      = errors.New("outbound: embedding_client request failed")
	ErrEmbeddingClientResponse     = errors.New("outbound: embedding_client response error")
)

// embeddingRequest represents the request payload for the embedding API.
type embeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

// embeddingResponse represents the response from the embedding API.
type embeddingResponse struct {
	Error *apiError       `json:"error,omitempty"`
	Data  []embeddingData `json:"data"`
}

// embeddingData represents a single embedding result.
type embeddingData struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

// apiError represents an error response from the API.
type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// EmbeddingClient is an implementation of the extraction.EmbeddingClient interface.
type EmbeddingClient struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	model      string
}

// NewEmbeddingClient creates a new instance of EmbeddingClient.
func NewEmbeddingClient(apiKey, baseURL, model string) (*EmbeddingClient, error) {
	if apiKey == "" {
		return nil, ErrEmbeddingClientEmptyAPIKey
	}
	if baseURL == "" {
		return nil, ErrEmbeddingClientEmptyBaseURL
	}
	if model == "" {
		return nil, ErrEmbeddingClientEmptyModel
	}

	return &EmbeddingClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiKey:     apiKey,
		baseURL:    baseURL,
		model:      model,
	}, nil
}

// Embed generates an embedding for the given text.
func (a *EmbeddingClient) Embed(note extraction.MemoryNote) (extraction.EmbeddedNote, error) {
	if note.Content == "" {
		return extraction.EmbeddedNote{}, ErrEmbeddingClientEmptyText
	}

	embedding, err := a.requestEmbedding(string(note.Content))
	if err != nil {
		return extraction.EmbeddedNote{}, err
	}

	return extraction.EmbeddedNote{
		Embedding: embedding,
		Note:      note,
	}, nil
}

// requestEmbedding sends a request to the embedding API and returns the embedding vector.
func (a *EmbeddingClient) requestEmbedding(text string) ([]float32, error) {
	reqBody := embeddingRequest{
		Input: text,
		Model: a.model,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrEmbeddingClientRequest, err)
	}

	req, err := http.NewRequest(http.MethodPost, a.baseURL+"/embeddings", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrEmbeddingClientRequest, err)
	}

	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrEmbeddingClientRequest, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrEmbeddingClientResponse, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d: %s", ErrEmbeddingClientResponse, resp.StatusCode, string(body))
	}

	var embResp embeddingResponse
	if err := json.Unmarshal(body, &embResp); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrEmbeddingClientResponse, err)
	}

	if embResp.Error != nil {
		return nil, fmt.Errorf("%w: %s", ErrEmbeddingClientResponse, embResp.Error.Message)
	}

	if len(embResp.Data) == 0 {
		return nil, fmt.Errorf("%w: no embedding data returned", ErrEmbeddingClientResponse)
	}

	return embResp.Data[0].Embedding, nil
}
