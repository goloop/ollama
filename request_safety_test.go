package ollama

import (
	"context"
	"errors"
	"github.com/goloop/ai"
	"testing"
)

func TestNativeChatRejectsNil(t *testing.T) {
	c := New("")
	if _, err := c.ChatCompletion(context.Background(), nil); !errors.Is(err, ai.ErrNoRequest) {
		t.Fatalf("got %v", err)
	}
	for _, err := range c.ChatStream(context.Background(), nil) {
		if !errors.Is(err, ai.ErrNoRequest) {
			t.Fatalf("got %v", err)
		}
	}
}

func TestDecodeStreamResponseRejectsErrors(t *testing.T) {
	for _, data := range []string{`{bad`, `{"error":"model unavailable"}`} {
		if _, err := decodeStreamResponse([]byte(data)); err == nil {
			t.Fatalf("accepted %s", data)
		}
	}
	got, err := decodeStreamResponse([]byte(`{"message":{"content":"hello"},"done":true}`))
	if err != nil || !got.Done || got.Message.Content != "hello" {
		t.Fatalf("got %+v %v", got, err)
	}
}
