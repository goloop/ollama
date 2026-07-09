package ollama

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestShow(t *testing.T) {
	c, done := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/show" {
			t.Errorf("path = %s", r.URL.Path)
		}
		var req map[string]string
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatal(err)
		}
		if req["model"] != "llama3" {
			t.Errorf("model = %q", req["model"])
		}
		io.WriteString(w, `{"template":"{{ .Prompt }}","parameters":"stop \"x\"",`+
			`"details":{"family":"llama","parameter_size":"8B","quantization_level":"Q4_0"},`+
			`"model_info":{"llama.context_length":8192},"capabilities":["completion","tools"]}`)
	})
	defer done()

	info, err := c.Show(context.Background(), "llama3")
	if err != nil {
		t.Fatal(err)
	}
	if info.Details.ParameterSize != "8B" {
		t.Errorf("parameter size = %q", info.Details.ParameterSize)
	}
	if info.ModelInfo["llama.context_length"] != float64(8192) {
		t.Errorf("context length = %v", info.ModelInfo["llama.context_length"])
	}
	if len(info.Capabilities) != 2 || info.Capabilities[0] != "completion" {
		t.Errorf("capabilities = %v", info.Capabilities)
	}
}

func TestShowError(t *testing.T) {
	c, done := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, `{"error":"model not found"}`)
	})
	defer done()

	_, err := c.Show(context.Background(), "nope")
	if err == nil {
		t.Fatal("want error")
	}
}
