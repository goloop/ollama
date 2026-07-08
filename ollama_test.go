package ollama

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goloop/ai"
)

func newTestClient(t *testing.T, h http.HandlerFunc) (*Client, func()) {
	t.Helper()
	srv := httptest.NewServer(h)
	c := New("", WithBaseURL(srv.URL), WithMaxRetries(0))
	return c, srv.Close
}

func TestGenerateText(t *testing.T) {
	c, done := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/api/chat") {
			t.Errorf("path = %q", r.URL.Path)
		}
		var req ChatRequest
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &req)
		if req.Stream {
			t.Error("stream should be false")
		}
		io.WriteString(w, `{"model":"m","message":{"role":"assistant","content":"hello"},`+
			`"done":true,"done_reason":"stop","prompt_eval_count":3,"eval_count":2}`)
	})
	defer done()

	resp, err := c.Generate(context.Background(), &ai.Request{
		Model:    "m",
		Messages: []ai.Message{ai.UserText("hi")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Text() != "hello" {
		t.Errorf("text = %q", resp.Text())
	}
	if resp.Usage.InputTokens != 3 || resp.Usage.OutputTokens != 2 {
		t.Errorf("usage = %+v", resp.Usage)
	}
	if resp.StopReason != "stop" {
		t.Errorf("stop = %q", resp.StopReason)
	}
}

func TestGenerateToolUse(t *testing.T) {
	c, done := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"model":"m","message":{"role":"assistant","content":"",`+
			`"tool_calls":[{"function":{"name":"get_weather","arguments":{"city":"Kyiv"}}}]},`+
			`"done":true}`)
	})
	defer done()

	resp, err := c.Generate(context.Background(), &ai.Request{
		Model:    "m",
		Messages: []ai.Message{ai.UserText("weather?")},
		Tools:    []ai.Tool{{Name: "get_weather"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	calls := resp.ToolCalls()
	if len(calls) != 1 || calls[0].Name != "get_weather" {
		t.Fatalf("calls = %+v", calls)
	}
	if string(calls[0].Input) != `{"city":"Kyiv"}` {
		t.Errorf("input = %s", calls[0].Input)
	}
}

func TestStream(t *testing.T) {
	lines := []string{
		`{"model":"m","message":{"role":"assistant","content":"Hel"},"done":false}`,
		`{"model":"m","message":{"role":"assistant","content":"lo"},"done":false}`,
		`{"model":"m","message":{"role":"assistant","content":""},"done":true,` +
			`"prompt_eval_count":5,"eval_count":2}`,
	}
	c, done := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		for _, line := range lines {
			io.WriteString(w, line+"\n")
		}
	})
	defer done()

	var text strings.Builder
	var usage ai.Usage
	var doneSeen bool
	for chunk, err := range c.Stream(context.Background(), &ai.Request{
		Model: "m", Messages: []ai.Message{ai.UserText("hi")},
	}) {
		if err != nil {
			t.Fatal(err)
		}
		text.WriteString(chunk.Text)
		if chunk.Done {
			doneSeen = true
			if chunk.Usage != nil {
				usage = *chunk.Usage
			}
		}
	}
	if text.String() != "Hello" || !doneSeen {
		t.Errorf("text = %q done = %v", text.String(), doneSeen)
	}
	if usage.InputTokens != 5 || usage.OutputTokens != 2 {
		t.Errorf("usage = %+v", usage)
	}
}

func TestStreamToolCall(t *testing.T) {
	lines := []string{
		`{"message":{"role":"assistant","content":"",` +
			`"tool_calls":[{"function":{"name":"lookup","arguments":{"q":42}}}]},"done":false}`,
		`{"message":{"role":"assistant","content":""},"done":true}`,
	}
	c, done := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		for _, line := range lines {
			io.WriteString(w, line+"\n")
		}
	})
	defer done()

	var call *ai.ToolUse
	for chunk, err := range c.Stream(context.Background(), &ai.Request{
		Model: "m", Messages: []ai.Message{ai.UserText("hi")},
	}) {
		if err != nil {
			t.Fatal(err)
		}
		if chunk.ToolCall != nil {
			call = chunk.ToolCall
		}
	}
	if call == nil || call.Name != "lookup" || string(call.Input) != `{"q":42}` {
		t.Fatalf("call = %+v", call)
	}
}

func TestError(t *testing.T) {
	c, done := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, `{"error":"model 'm' not found"}`)
	})
	defer done()

	_, err := c.Generate(context.Background(), &ai.Request{
		Model: "m", Messages: []ai.Message{ai.UserText("hi")},
	})
	var apiErr *ai.APIError
	if !errors.As(err, &apiErr) || apiErr.Status != 404 ||
		apiErr.Message != "model 'm' not found" {
		t.Fatalf("err = %v", err)
	}
}

func TestMessagesMapping(t *testing.T) {
	req := &ai.Request{
		System: "sys",
		Messages: []ai.Message{
			{Role: ai.RoleUser, Parts: []ai.Part{
				ai.Text{Text: "look"},
				ai.Image{MIME: "image/png", Data: []byte{1, 2, 3}},
			}},
			{Role: ai.RoleAssistant, Parts: []ai.Part{
				ai.ToolUse{ID: "t", Name: "t", Input: json.RawMessage(`{}`)},
			}},
			{Role: ai.RoleTool, Parts: []ai.Part{
				ai.ToolResult{ID: "t", Content: "42"},
			}},
		},
	}
	msgs := ollamaMessages(req)
	if msgs[0].Role != "system" || msgs[0].Content != "sys" {
		t.Errorf("system = %+v", msgs[0])
	}
	if msgs[1].Role != "user" || len(msgs[1].Images) != 1 {
		t.Fatalf("user = %+v", msgs[1])
	}
	if msgs[1].Images[0] != base64.StdEncoding.EncodeToString([]byte{1, 2, 3}) {
		t.Errorf("image = %q", msgs[1].Images[0])
	}
	if len(msgs[2].ToolCalls) != 1 || msgs[2].ToolCalls[0].Function.Name != "t" {
		t.Errorf("assistant = %+v", msgs[2])
	}
	if msgs[3].Role != "tool" || msgs[3].Content != "42" {
		t.Errorf("tool = %+v", msgs[3])
	}
}

func TestValidate(t *testing.T) {
	c := New("")
	_, err := c.Generate(context.Background(), &ai.Request{Model: "m"})
	if !errors.Is(err, ai.ErrNoMessages) {
		t.Errorf("want ErrNoMessages, got %v", err)
	}
}
