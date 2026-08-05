[![deps.dev](https://img.shields.io/badge/deps.dev-insights-4c8dbc)](https://deps.dev/go/github.com%2Fgoloop%2Follama) [![License](https://img.shields.io/badge/license-MIT-brightgreen)](https://github.com/goloop/ollama/blob/master/LICENSE) [![License](https://img.shields.io/badge/godoc-YES-green)](https://pkg.go.dev/github.com/goloop/ollama) [![Stay with Ukraine](https://img.shields.io/static/v1?label=Stay%20with&message=Ukraine%20♥&color=ffD700&labelColor=0057B8&style=flat)](https://u24.gov.ua/)


# ollama

`ollama` is a Go client for a local or remote [Ollama](https://ollama.com)
server. It implements the `github.com/goloop/ai` interface, so it looks and
works like every other goloop AI provider, and exposes Ollama's native
endpoints on top.

## Features

- Chat: `Generate` for a single response, `Stream` for token-by-token output
  through `iter.Seq2`.
- Tool use (function calling) and image input.
- Native `ChatCompletion` and `ChatStream` over `/api/chat`.
- Embeddings, the installed-model list and per-model details (`Show`).
- Retries on 429 and 5xx with backoff; normalized, typed API errors.
- Depends only on `github.com/goloop/ai` and the standard library.
- Structured output: `ai.Format` maps onto the server's `format` field
  (`"json"` or a schema); read the reply with `resp.JSON(&v)`.

## Installation

```sh
go get github.com/goloop/ollama
```

## Quick start

```go
package main

import (
	"context"
	"fmt"

	"github.com/goloop/ai"
	"github.com/goloop/ollama"
)

func main() {
	c := ollama.New("") // local server on http://localhost:11434

	resp, err := c.Generate(context.Background(), &ai.Request{
		Model:    ollama.ModelLlama32,
		Messages: []ai.Message{ai.UserText("Say hello in one word.")},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Text())
}
```

## Remote server

```go
c := ollama.New(apiKey, ollama.WithBaseURL("https://ollama.example.com"))
```

The API key is optional; it is sent as a bearer token when set, for
authenticated proxies in front of a server.

## Streaming

Ollama streams newline-delimited JSON rather than Server-Sent Events; `Stream`
hides that difference and yields the same `ai.Chunk` values as every other
provider.

```go
for chunk, err := range c.Stream(ctx, req) {
	if err != nil {
		break
	}
	fmt.Print(chunk.Text)
}
```

## Embeddings and models

```go
vecs, err := c.Embed(ctx, ollama.ModelLlama32, "hello", "world")
models, err := c.Models(ctx) // what is installed locally
info, err := c.Show(ctx, ollama.ModelLlama32) // template, params, capabilities
```

## Documentation

Full reference: **[DOC.md](DOC.md)** (Ukrainian: **[DOC.UK.md](DOC.UK.md)**).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT - see [LICENSE](LICENSE).
