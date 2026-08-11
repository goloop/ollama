package ollama

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goloop/ai"
)

func askToSearch() *ai.Request {
	return &ai.Request{
		Model:    "the-model",
		Messages: []ai.Message{ai.UserText("What happened today?")},
		Hosted:   []ai.Hosted{{Kind: ai.HostedWebSearch}},
	}
}

// This provider cannot run a hosted capability, and that is a documented
// answer rather than an oversight: the request is refused before it leaves,
// so a caller never gets an answer that only looks researched.
func TestHostedIsRefusedBeforeTheRequestLeaves(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			called = true
			_, _ = w.Write([]byte(`{}`))
		}))
	defer srv.Close()

	c := New("k", WithBaseURL(srv.URL))

	if _, err := c.Generate(context.Background(), askToSearch()); !errors.Is(
		err, ai.ErrNoHosted) {
		t.Errorf("Generate() error = %v, want ErrNoHosted", err)
	}

	var streamErr error
	for _, err := range c.Stream(context.Background(), askToSearch()) {
		if err != nil {
			streamErr = err
		}
	}
	if !errors.Is(streamErr, ai.ErrNoHosted) {
		t.Errorf("Stream() error = %v, want ErrNoHosted", streamErr)
	}

	if called {
		t.Error("a request the provider cannot serve was sent anyway")
	}
}

// The other half: a request that asks for nothing hosted is untouched and
// still goes out exactly as it did before.
func TestWithoutHostedTheRequestStillGoesOut(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			called = true
			_, _ = w.Write([]byte(`{}`))
		}))
	defer srv.Close()

	req := askToSearch()
	req.Hosted = nil

	// The reply is deliberately empty: what is being checked is that the
	// request left, not what came back.
	_, _ = New("k", WithBaseURL(srv.URL)).Generate(context.Background(), req)
	if !called {
		t.Error("a plain request never reached the provider")
	}
}
