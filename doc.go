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
// # Hosted capabilities
//
// Models here run on the machine that serves them, and that server has no
// search to run. ai.Hosted is refused with ai.ErrNoHosted.
//
// The refusal is the documented behavior, not a gap waiting to be filled
// in silence: an answer produced without the search that was asked for
// looks exactly like one produced with it. A caller who would rather have
// the answer anyway asks again without ai.Request.Hosted.
//
// # Asking what this driver can do
//
// Capabilities describes this driver for the decision taken before a call:
// whether to offer a feature at all, and whether it needs one request or two.
//
//	if ai.SupportsHosted(c, ai.Hosted{Kind: ai.HostedWebSearch}) { ... }
//
// It is a hint and not a permission - support also depends on the model, the
// account and the region - so ai.ErrNoHosted and ai.ErrNoFormat remain the
// source of truth and a caller still handles them. What changes is that a
// refusal the provider only reports as a 400 now arrives as those same
// sentinels, wrapped around the original ai.APIError, so one errors.Is covers
// a limitation this driver knew in advance and one it learned over the wire.
//
// It depends only on goloop/ai and the standard library.
package ollama
