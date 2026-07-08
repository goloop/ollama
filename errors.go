package ollama

import (
	"encoding/json"

	"github.com/goloop/ai"
)

// parseError turns a non-success response body into an *ai.APIError. Ollama
// reports errors as {"error":"message"}.
func parseError(status int, body []byte) error {
	var w struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(body, &w)

	return &ai.APIError{
		Status:  status,
		Message: w.Error,
		Raw:     append(json.RawMessage(nil), body...),
	}
}
