package trussium

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompleteForwardsRequestID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" || r.Header.Get("X-Request-ID") != "request-123" {
			t.Fatal("unexpected request")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"chat-1","provider":"ollama","model":"llama","choices":[{"index":0,"message":{"role":"assistant","content":"hello"},"finish_reason":"stop"}],"usage":{"input_tokens":1,"output_tokens":1,"total_tokens":2}}`))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Complete(context.Background(), ChatCompletionRequest{Model: "llama", Messages: []Message{{Role: "user", Content: "hello"}}}, "request-123")
	if err != nil || result.Choices[0].Message.Content != "hello" {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestReadinessCapabilitiesAndAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/health/ready":
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case "/v1/capabilities":
			_, _ = w.Write([]byte(`{"capabilities":[{"name":"chat.completions"}]}`))
		default:
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error":{"code":"provider_unreachable"}}`))
		}
	}))
	defer server.Close()
	client, err := NewClient(server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	ready, err := client.Readiness(context.Background())
	if err != nil || ready.Status != "ok" {
		t.Fatal(err)
	}
	caps, err := client.Capabilities(context.Background())
	if err != nil || caps.Capabilities[0].Name != "chat.completions" {
		t.Fatal(err)
	}
	err = client.do(context.Background(), http.MethodGet, "/failure", nil, "", &struct{}{})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "provider_unreachable" {
		t.Fatalf("error=%v", err)
	}
}
