package ollama

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/goloop/ai"
)

func askFor(f *ai.Format) *ai.Request {
	return &ai.Request{
		Model:    "the-model",
		Messages: []ai.Message{ai.UserText("Describe this article.")},
		Format:   f,
	}
}

// TestResponseFormat pins this provider's shape: the format field takes the
// bare word "json" or a schema directly, with no wrapper object.
func TestResponseFormat(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"a":{"type":"string"}}}`)

	cases := []struct {
		name   string
		format *ai.Format
		want   string
		mode   ai.FormatMode
	}{
		{"nothing asked", nil, "", ai.FormatNone},
		{"text", &ai.Format{Type: ai.FormatText}, "", ai.FormatNone},
		{"json", &ai.Format{Type: ai.FormatJSON}, `"json"`, ai.FormatNative},
		{
			"json schema",
			&ai.Format{Type: ai.FormatJSONSchema, Schema: schema},
			string(schema), ai.FormatNative,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cr, err := (&Client{}).chatRequest(askFor(c.format), false)
			if err != nil {
				t.Fatal(err)
			}
			if got := string(cr.Format); got != c.want {
				t.Errorf("format = %s, want %s", got, c.want)
			}
			if got := formatMode(c.format); got != c.mode {
				t.Errorf("formatMode = %s, want %s", got, c.mode)
			}
		})
	}

	_, err := (&Client{}).chatRequest(askFor(&ai.Format{Type: ai.FormatType(99)}), false)
	if !errors.Is(err, ai.ErrBadFormat) {
		t.Errorf("err = %v, want %v", err, ai.ErrBadFormat)
	}
}
