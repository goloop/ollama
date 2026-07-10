package ollama

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/goloop/ai"
)

// TestStreamTruncated verifies a stream that ends without a final done object
// surfaces io.ErrUnexpectedEOF instead of silently emitting a done chunk.
func TestStreamTruncated(t *testing.T) {
	c, done := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// Two content chunks, then the connection ends: no "done":true.
		io.WriteString(w, `{"message":{"content":"Hel"},"done":false}`+"\n")
		io.WriteString(w, `{"message":{"content":"lo"},"done":false}`+"\n")
	})
	defer done()

	var gotErr error
	var text string
	for chunk, err := range c.Stream(context.Background(), &ai.Request{
		Model: "m", Messages: []ai.Message{ai.UserText("hi")},
	}) {
		if err != nil {
			gotErr = err
			break
		}
		text += chunk.Text
	}
	if !errors.Is(gotErr, io.ErrUnexpectedEOF) {
		t.Fatalf("err = %v, want io.ErrUnexpectedEOF", gotErr)
	}
	if text != "Hello" {
		t.Errorf("text = %q", text)
	}
}
