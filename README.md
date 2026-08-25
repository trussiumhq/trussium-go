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

## Runnable self-hosted example

With a Trussium runtime running on port 9000, run the example from a source
checkout:

```bash
TRUSSIUM_URL=http://127.0.0.1:9000 \
TRUSSIUM_MODEL=llama3.1:8b \
TRUSSIUM_PROMPT="Say hello." \
go run ./examples/basic
```

The environment variables are optional. The example checks readiness and
capabilities before making one completion request. It calls an existing local,
private, or public runtime; it does not install or host Trussium.

## Development

```bash
gofmt -w *.go
go vet ./...
go test ./...
```
