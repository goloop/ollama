package ollama

import "github.com/goloop/ai"

// DefaultBaseURL is the base URL of a local Ollama server.
const DefaultBaseURL = "http://localhost:11434"

// Convenience model identifiers. Any model tag installed on the server is
// accepted; use Models to list what is available locally.
const (
	ModelLlama32 = "llama3.2"
	ModelLlama31 = "llama3.1"
	ModelMistral = "mistral"
	ModelGemma3  = "gemma3"
	ModelQwen25  = "qwen2.5"
)

// Client is an Ollama API client. It implements [ai.Client] and adds the
// provider's native endpoints. It talks to a local (or remote) Ollama server
// using its native API.
type Client struct {
	opts ai.Options
}

var _ ai.Client = (*Client)(nil)

// New returns a Client. The API key may be empty for a local server; when set,
// it is sent as a bearer token for authenticated proxies. Shared options
// (WithBaseURL, WithHTTPClient, WithTimeout, WithMaxRetries, WithHeader)
// configure it.
func New(apiKey string, opts ...Option) *Client {
	s := settings{}
	for _, o := range opts {
		o(&s)
	}

	o := ai.NewOptions(apiKey, s.aiOpts...)
	if o.BaseURL == "" {
		o.BaseURL = DefaultBaseURL
	}

	return &Client{opts: o}
}
