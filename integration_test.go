//go:build integration

// Integration smoke tests hit a live Ollama server. They are excluded from the
// normal build and, because Ollama runs locally without a key, they run only
// when OLLAMA_TEST is set (so they never hit localhost by accident):
//
//	OLLAMA_TEST=1 go test -tags integration -run Integration ./...
//
// Set OLLAMA_HOST to point at a non-default server. The model in OLLAMA_MODEL
// (default llama3.2) must already be pulled.
package ollama_test

import (
	"context"
	"os"
	"testing"

	"github.com/goloop/ai"
	"github.com/goloop/ollama"
)

func integrationClient(t *testing.T) (*ollama.Client, string) {
	t.Helper()
	if os.Getenv("OLLAMA_TEST") == "" {
		t.Skip("set OLLAMA_TEST=1 to run integration tests against a local server")
	}
	var opts []ollama.Option
	if host := os.Getenv("OLLAMA_HOST"); host != "" {
		opts = append(opts, ollama.WithBaseURL(host))
	}
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = ollama.ModelLlama32
	}
	return ollama.New("", opts...), model
}

func TestIntegrationGenerate(t *testing.T) {
	c, model := integrationClient(t)
	resp, err := c.Generate(context.Background(), &ai.Request{
		Model:    model,
		Messages: []ai.Message{ai.UserText("Reply with exactly one word: pong")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Text() == "" {
		t.Fatal("empty text")
	}
	t.Logf("generate: %q (in=%d out=%d)", resp.Text(), resp.Usage.InputTokens, resp.Usage.OutputTokens)
}

func TestIntegrationStream(t *testing.T) {
	c, model := integrationClient(t)
	var text string
	var done bool
	for chunk, err := range c.Stream(context.Background(), &ai.Request{
		Model:    model,
		Messages: []ai.Message{ai.UserText("Count from 1 to 5.")},
	}) {
		if err != nil {
			t.Fatal(err)
		}
		text += chunk.Text
		if chunk.Done {
			done = true
		}
	}
	if text == "" || !done {
		t.Fatalf("text=%q done=%v", text, done)
	}
	t.Logf("stream: %q done=%v", text, done)
}

func TestIntegrationShow(t *testing.T) {
	c, model := integrationClient(t)
	info, err := c.Show(context.Background(), model)
	if err != nil {
		t.Fatal(err)
	}
	if info.Template == "" && len(info.ModelInfo) == 0 {
		t.Fatal("Show returned no template and no model_info")
	}
	t.Logf("show: caps=%v info_keys=%d", info.Capabilities, len(info.ModelInfo))
}

func TestIntegrationEmbed(t *testing.T) {
	c, model := integrationClient(t)
	vecs, err := c.Embed(context.Background(), model, "hello", "world")
	if err != nil {
		t.Fatal(err)
	}
	if len(vecs) != 2 || len(vecs[0]) == 0 {
		t.Fatalf("vectors = %d x %d", len(vecs), len(vecs[0]))
	}
	t.Logf("embed: %d vectors of dim %d", len(vecs), len(vecs[0]))
}
