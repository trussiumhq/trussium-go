# Trussium Go SDK

The official Go SDK calls an existing Trussium runtime. It does not install or
host the runtime, Kubernetes, Helm, or `trussium-operator`.

## Install

```bash
go get github.com/trussiumhq/trussium-go
```

## Usage

```go
client, err := trussium.NewClient("http://127.0.0.1:9000", nil)
response, err := client.Complete(context.Background(), trussium.ChatCompletionRequest{
    Model: "llama3.1:8b",
    Messages: []trussium.Message{{Role: "user", Content: "Say hello."}},
}, "request-123")
```

The foundation provides context-aware non-streaming chat completions,
`GET /health/ready`, and `GET /v1/capabilities`. A caller-supplied request ID
is forwarded as `X-Request-ID`; non-success responses return `*APIError`.

## Development

```bash
gofmt -w *.go
go vet ./...
go test ./...
```
