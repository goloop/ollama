package ollama

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/goloop/ai"
)

// A 400 that names a capability the caller asked for is the provider saying
// the model cannot do it. Both readings must survive: errors.Is so the
// application can degrade with one check, errors.As so the provider's own
// message is still there to log.
func TestWrapUnsupportedKeepsBothReadings(t *testing.T) {
	original := &ai.APIError{
		Status:  http.StatusBadRequest,
		Message: "response_format is not supported by this model",
		Raw:     json.RawMessage(`{"error":{"param":"response_format"}}`),
	}

	req := &ai.Request{
		Model:    "the-model",
		Messages: []ai.Message{ai.UserText("hi")},
		Format:   &ai.Format{Type: ai.FormatJSON},
	}

	err := wrapUnsupportedCapability(req, original)
	if !errors.Is(err, ai.ErrNoFormat) {
		t.Fatalf("wrapUnsupportedCapability() = %v, want ai.ErrNoFormat", err)
	}

	var apiErr *ai.APIError
	if !errors.As(err, &apiErr) {
		t.Fatal("the provider's APIError was lost")
	}
	if apiErr.Status != http.StatusBadRequest || apiErr.Message != original.Message {
		t.Errorf("APIError = %+v, want the original", apiErr)
	}
}

// The narrow rules matter more than the translation: a false positive turns a
// genuinely bad request into an endless "the provider cannot do this" retry.
func TestWrapUnsupportedIsNarrow(t *testing.T) {
	withFormat := func() *ai.Request {
		return &ai.Request{
			Model:    "the-model",
			Messages: []ai.Message{ai.UserText("hi")},
			Format:   &ai.Format{Type: ai.FormatJSON},
		}
	}
	plain := func() *ai.Request {
		return &ai.Request{
			Model:    "the-model",
			Messages: []ai.Message{ai.UserText("hi")},
		}
	}

	tests := []struct {
		name string
		req  *ai.Request
		err  error
	}{
		{
			name: "not a 400",
			req:  withFormat(),
			err: &ai.APIError{
				Status:  http.StatusInternalServerError,
				Message: "response_format is not supported by this model",
			},
		},
		{
			name: "the caller never asked for it",
			req:  plain(),
			err: &ai.APIError{
				Status:  http.StatusBadRequest,
				Message: "response_format is not supported by this model",
			},
		},
		{
			name: "a 400 about something else",
			req:  withFormat(),
			err: &ai.APIError{
				Status:  http.StatusBadRequest,
				Message: "messages: at least one message is required",
			},
		},
		{
			name: "names the feature but is not a refusal",
			req:  withFormat(),
			err: &ai.APIError{
				Status:  http.StatusBadRequest,
				Message: "response_format must be an object",
			},
		},
		{
			name: "not an APIError at all",
			req:  withFormat(),
			err:  errors.New("dial tcp: connection refused"),
		},
		{
			name: "no request to check against",
			req:  nil,
			err: &ai.APIError{
				Status:  http.StatusBadRequest,
				Message: "response_format is not supported by this model",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapUnsupportedCapability(tt.req, tt.err)
			if errors.Is(got, ai.ErrNoFormat) || errors.Is(got, ai.ErrNoHosted) {
				t.Errorf("wrapUnsupportedCapability() translated %v", tt.err)
			}
			if got != tt.err {
				t.Errorf("the error was replaced: %v", got)
			}
		})
	}
}

// A structured field beats prose: it survives rewording and localisation.
func TestWrapUnsupportedPrefersStructuredFields(t *testing.T) {
	req := &ai.Request{
		Model:    "the-model",
		Messages: []ai.Message{ai.UserText("hi")},
		Format:   &ai.Format{Type: ai.FormatJSON},
	}

	// A message in a language nobody matched, with the blamed parameter named
	// in the body where the provider reports it.
	err := wrapUnsupportedCapability(req, &ai.APIError{
		Status:  http.StatusBadRequest,
		Message: "непідтримуване значення",
		Raw:     json.RawMessage(`{"error":{"param":"response_format"}}`),
	})
	if !errors.Is(err, ai.ErrNoFormat) {
		t.Errorf("a named parameter was ignored in favour of prose: %v", err)
	}
}

// This driver runs nothing of its own, and its Capabilities must say so rather
// than leave a caller guessing from an empty map it might not have checked.
func TestCapabilitiesReportNoHosted(t *testing.T) {
	caps := (&Client{}).Capabilities()
	if len(caps.Hosted) != 0 {
		t.Errorf("Hosted = %+v, want empty for a driver with no hosted "+
			"capability", caps.Hosted)
	}

	req := &ai.Request{
		Model:    "the-model",
		Messages: []ai.Message{ai.UserText("hi")},
		Hosted:   []ai.Hosted{{Kind: ai.HostedWebSearch}},
	}
	if err := checkHosted(req); !errors.Is(err, ai.ErrNoHosted) {
		t.Errorf("checkHosted() = %v, want ai.ErrNoHosted - the table and "+
			"the behaviour must agree", err)
	}
	if ai.SupportsHosted(&Client{}, req.Hosted[0]) {
		t.Error("ai.SupportsHosted() said yes for a driver that refuses it")
	}
}

// The format modes reported must match what the driver actually reports on a
// response, or the hint is worse than none.
func TestCapabilitiesMatchFormatMode(t *testing.T) {
	caps := (&Client{}).Capabilities()
	if got := formatMode(&ai.Format{Type: ai.FormatJSON}); got != caps.Format.JSON {
		t.Errorf("formatMode(JSON) = %v but Capabilities says %v",
			got, caps.Format.JSON)
	}
	schema := &ai.Format{Type: ai.FormatJSONSchema}
	if got := formatMode(schema); got != caps.Format.JSONSchema {
		t.Errorf("formatMode(JSONSchema) = %v but Capabilities says %v",
			got, caps.Format.JSONSchema)
	}
}
