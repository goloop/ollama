// Package ollama is a client for a local or remote Ollama server, built on the
// goloop/ai interface.
//
// The Client implements ai.Client, so Generate and Stream work the same as
// with any other goloop AI provider. On top of that it exposes Ollama's native
// endpoints: /api/chat (with tool use and image input), embeddings, the
// installed-model list and per-model details (Show). Ollama streams
// newline-delimited JSON rather than
// Server-Sent Events; Stream hides that difference.
//
//	c := ollama.New("") // local server, no API key
//	resp, err := c.Generate(ctx, &ai.Request{
//	    Model:    ollama.ModelLlama32,
//	    Messages: []ai.Message{ai.UserText("Say hello in one word.")},
//	})
//
// # Structured output
//
// ai.Request.Format maps onto the server's format field - the bare word "json",
// or a schema directly - so a request for JSON is enforced rather than merely
// asked for, and ai.Response.JSON decodes the reply.
//
// It depends only on goloop/ai and the standard library.
package ollama
