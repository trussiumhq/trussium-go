// Package trussium calls an existing Trussium runtime over HTTP.
package trussium

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client calls one configured Trussium runtime.
type Client struct {
	baseURL string
	http    *http.Client
}

// APIError describes a non-success response from the runtime.
type APIError struct {
	StatusCode int
	Code       string
}

func (err *APIError) Error() string {
	if err.Code == "" {
		return fmt.Sprintf("trussium runtime returned HTTP %d", err.StatusCode)
	}
	return fmt.Sprintf("trussium runtime returned HTTP %d: %s", err.StatusCode, err.Code)
}

// NewClient creates a client for baseURL. An empty URL uses the local runtime.
// Supply an http.Client to control transport, timeout, or authentication.
func NewClient(baseURL string, httpClient *http.Client) (*Client, error) {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:9000"
	}
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		return nil, errors.New("trussium base URL must use http or https")
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), http: httpClient}, nil
}

// Message is one normalized chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest is a non-streaming chat-completion request.
type ChatCompletionRequest struct {
	Model           string    `json:"model"`
	Messages        []Message `json:"messages"`
	Temperature     *float64  `json:"temperature,omitempty"`
	MaxOutputTokens *int      `json:"max_output_tokens,omitempty"`
}

// TokenUsage reports normalized token counts.
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// ChatChoice is one completed chat result.
type ChatChoice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// ChatCompletion is a completed non-streaming chat response.
type ChatCompletion struct {
	ID       string       `json:"id"`
	Provider string       `json:"provider"`
	Model    string       `json:"model"`
	Choices  []ChatChoice `json:"choices"`
	Usage    TokenUsage   `json:"usage"`
}

type TranslationRequest struct {
	Model          string   `json:"model"`
	Input          []string `json:"input"`
	SourceLanguage string   `json:"source_language,omitempty"`
	TargetLanguage string   `json:"target_language"`
}

type TranslationResult struct {
	Text           string `json:"text"`
	SourceLanguage string `json:"source_language,omitempty"`
	TargetLanguage string `json:"target_language"`
}
type TranslationResponse struct {
	ID           string              `json:"id"`
	Provider     string              `json:"provider"`
	Model        string              `json:"model"`
	Translations []TranslationResult `json:"translations"`
}

// Readiness is the runtime traffic-readiness response.
type Readiness struct {
	Status string `json:"status"`
}

// Capability is public metadata for one configured capability.
type Capability struct {
	Name string `json:"name"`
}

// Capabilities is the stable capability-discovery response.
type Capabilities struct {
	Capabilities []Capability `json:"capabilities"`
}

// Complete performs one non-streaming chat completion. A non-empty requestID is
// forwarded as X-Request-ID for caller-controlled correlation.
func (client *Client) Complete(ctx context.Context, payload ChatCompletionRequest, requestID string) (ChatCompletion, error) {
	var result ChatCompletion
	if err := client.do(ctx, http.MethodPost, "/v1/chat/completions", payload, requestID, &result); err != nil {
		return ChatCompletion{}, err
	}
	return result, nil
}

// Readiness returns the runtime traffic-readiness state.
func (client *Client) Readiness(ctx context.Context) (Readiness, error) {
	var result Readiness
	if err := client.do(ctx, http.MethodGet, "/health/ready", nil, "", &result); err != nil {
		return Readiness{}, err
	}
	return result, nil
}

// Capabilities returns public metadata for configured capabilities.
func (client *Client) Capabilities(ctx context.Context) (Capabilities, error) {
	var result Capabilities
	if err := client.do(ctx, http.MethodGet, "/v1/capabilities", nil, "", &result); err != nil {
		return Capabilities{}, err
	}
	return result, nil
}

func (client *Client) Translate(ctx context.Context, payload TranslationRequest, requestID string) (TranslationResponse, error) {
	var result TranslationResponse
	if err := client.do(ctx, http.MethodPost, "/v1/translations", payload, requestID, &result); err != nil {
		return TranslationResponse{}, err
	}
	return result, nil
}

func (client *Client) do(ctx context.Context, method, path string, payload any, requestID string, result any) error {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode trussium request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, client.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("build trussium request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if requestID != "" {
		request.Header.Set("X-Request-ID", requestID)
	}
	response, err := client.http.Do(request)
	if err != nil {
		return fmt.Errorf("call trussium runtime: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return decodeAPIError(response)
	}
	if err := json.NewDecoder(response.Body).Decode(result); err != nil {
		return fmt.Errorf("decode trussium response: %w", err)
	}
	return nil
}

func decodeAPIError(response *http.Response) error {
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if json.NewDecoder(response.Body).Decode(&payload) != nil {
		return &APIError{StatusCode: response.StatusCode}
	}
	return &APIError{StatusCode: response.StatusCode, Code: payload.Error.Code}
}
