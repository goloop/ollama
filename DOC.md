# ollama - reference

The full reference for the `ollama` package: the client, the shared `goloop/ai`
model, chat (interface and native), streaming, embeddings and models.

Ukrainian version: **[DOC.UK.md](DOC.UK.md)**.

## Contents

- [Mental model](#mental-model)
- [Creating a client](#creating-a-client)
- [Generate and Stream](#generate-and-stream)
- [Native chat](#native-chat)
- [Tools and images](#tools-and-images)
- [Embeddings](#embeddings)
- [Models](#models)
- [Options and errors](#options-and-errors)

## Mental model

`ollama.Client` implements `ai.Client`, the provider-agnostic contract from
`github.com/goloop/ai`. The shared `Generate` and `Stream` cover the common
ground - chat with tools, images and streaming - so code written against the
interface runs on any provider, local or hosted.

Native power lives in `ChatCompletion`/`ChatStream` over `/api/chat`, embeddings
and the installed-model list. Ollama streams newline-delimited JSON, not
Server-Sent Events; the driver hides that.

```go
import (
	"github.com/goloop/ai"
	"github.com/goloop/ollama"
)
```

## Creating a client

```go
c := ollama.New("") // local server on http://localhost:11434

c = ollama.New(apiKey, ollama.WithBaseURL("https://ollama.example.com"))
```

The API key is optional; when set it is sent as a bearer token for
authenticated proxies in front of a server.

## Generate and Stream

```go
resp, err := c.Generate(ctx, &ai.Request{
	Model:    ollama.ModelLlama32,
	System:   "You are concise.",
	Messages: []ai.Message{ai.UserText("Name three primary colors.")},
})
resp.Text()
resp.ToolCalls()
resp.Usage
```

`Stream` returns `iter.Seq2[ai.Chunk, error]`: text deltas as chunks with
`Text`, a tool call as a chunk with `ToolCall`, and a final chunk with `Done`
and `Usage`.

```go
for chunk, err := range c.Stream(ctx, req) {
	if err != nil {
		return err
	}
	fmt.Print(chunk.Text)
}
```

If the stream ends before Ollama marks it done, `Stream` yields
`io.ErrUnexpectedEOF` rather than silently reporting a completed response.

## Native chat

For provider-only options build a `ChatRequest` and call `ChatCompletion` or
`ChatStream`:

```go
resp, err := c.ChatCompletion(ctx, &ollama.ChatRequest{
	Model:    ollama.ModelLlama32,
	Messages: []ollama.Message{{Role: "user", Content: "as JSON"}},
	Format:   json.RawMessage(`"json"`),
	Options:  &ollama.Options{NumPredict: 256},
})
```

`Options` maps to Ollama's generation options (`temperature`, `top_p`,
`num_predict`, `stop`); `Format` sets structured output.

## Tools and images

Tool use and images use the shared `ai` types: `ai.Tool`, `ai.Image`,
`ai.ToolResult`. Ollama matches tool results to calls positionally rather than
by an ID, so a returned `ai.ToolUse` carries the function name as its `ID`.
Inline image bytes are sent base64-encoded (URL images are not fetched by the
server).

## Embeddings

```go
vecs, err := c.Embed(ctx, ollama.ModelLlama32, "hello", "world")
```

## Models

```go
models, err := c.Models(ctx) // installed models (the /api/tags endpoint)
models[0].Name               // "llama3.2:latest"
models[0].Details.ParameterSize
```

## Options and errors

Options: `WithBaseURL`, `WithHTTPClient`, `WithTimeout`, `WithMaxRetries`,
`WithHeader`.

A non-success response becomes an `*ai.APIError` with `Status`, `Message` and
the raw body:

```go
var apiErr *ai.APIError
if errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound {
	// model not pulled yet
}
```

Requests missing a model or messages fail before the network with
`ai.ErrNoModel` or `ai.ErrNoMessages`.
