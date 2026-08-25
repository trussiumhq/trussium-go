package main

import (
	"context"
	"fmt"
	"os"

	trussium "github.com/trussiumhq/trussium-go"
)

func main() {
	baseURL := valueOrDefault("TRUSSIUM_URL", "http://127.0.0.1:9000")
	model := valueOrDefault("TRUSSIUM_MODEL", "llama3.1:8b")
	prompt := valueOrDefault("TRUSSIUM_PROMPT", "Say hello from the Go SDK.")

	client, err := trussium.NewClient(baseURL, nil)
	if err != nil {
		fatal(err)
	}
	ctx := context.Background()
	ready, err := client.Readiness(ctx)
	if err != nil {
		fatal(err)
	}
	fmt.Printf("Ready: %+v\n", ready)
	capabilities, err := client.Capabilities(ctx)
	if err != nil {
		fatal(err)
	}
	fmt.Printf("Capabilities: %+v\n", capabilities)
	response, err := client.Complete(ctx, trussium.ChatCompletionRequest{
		Model:    model,
		Messages: []trussium.Message{{Role: "user", Content: prompt}},
	}, "go-example")
	if err != nil {
		fatal(err)
	}
	fmt.Printf("Completion: %+v\n", response)
}

func valueOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "trussium example: %v\n", err)
	os.Exit(1)
}
