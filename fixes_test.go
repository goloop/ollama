package ollama

import (
	"context"
	"io"
	"net/http"
	"testing"
)

// Native methods must not mutate the caller's request (Stream field).
func TestFixChatRequestNotMutated(t *testing.T) {
	c, done := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"model":"m","message":{"role":"assistant","content":"hi"},"done":true}`)
	})
	defer done()

	req := &ChatRequest{Model: "m", Messages: []Message{{Role: "user", Content: "hi"}}}
	if _, err := c.ChatCompletion(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if req.Stream {
		t.Error("ChatCompletion mutated req.Stream")
	}
}
