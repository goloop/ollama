package ollama

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestChatCompletionNative(t *testing.T) {
	c, done := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"model":"m","message":{"role":"assistant","content":"hi"},"done":true}`)
	})
	defer done()

	resp, err := c.ChatCompletion(context.Background(), &ChatRequest{
		Model:    "m",
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	if err != nil || resp.Message.Content != "hi" {
		t.Fatalf("chat: %v %+v", err, resp)
	}
}

func TestChatStreamNative(t *testing.T) {
	lines := []string{
		`{"message":{"content":"a"},"done":false}`,
		`{"message":{"content":"b"},"done":true}`,
	}
	c, done := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		for _, line := range lines {
			io.WriteString(w, line+"\n")
		}
	})
	defer done()

	var text strings.Builder
	for chunk, err := range c.ChatStream(context.Background(), &ChatRequest{
		Model: "m", Messages: []Message{{Role: "user", Content: "hi"}},
	}) {
		if err != nil {
			t.Fatal(err)
		}
		text.WriteString(chunk.Message.Content)
	}
	if text.String() != "ab" {
		t.Errorf("text = %q", text.String())
	}
}

func TestEmbed(t *testing.T) {
	c, done := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/api/embed") {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{"embeddings":[[0.1,0.2],[0.3,0.4]]}`)
	})
	defer done()

	vecs, err := c.Embed(context.Background(), ModelLlama32, "a", "b")
	if err != nil || len(vecs) != 2 || vecs[1][0] != 0.3 {
		t.Fatalf("embed: %v %+v", err, vecs)
	}
}

func TestModels(t *testing.T) {
	c, done := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/api/tags") {
			t.Errorf("path = %q", r.URL.Path)
		}
		io.WriteString(w, `{"models":[{"name":"llama3.2:latest","model":"llama3.2",`+
			`"details":{"parameter_size":"3B"}}]}`)
	})
	defer done()

	models, err := c.Models(context.Background())
	if err != nil || len(models) != 1 || models[0].Name != "llama3.2:latest" {
		t.Fatalf("models: %v %+v", err, models)
	}
	if models[0].Details.ParameterSize != "3B" {
		t.Errorf("details = %+v", models[0].Details)
	}
}
