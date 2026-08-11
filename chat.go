package ollama

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/goloop/ai"
)

// Message is one message in a chat request or response. Images are base64-
// encoded; ToolCalls are set on assistant messages that call a tool.
type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	Images    []string   `json:"images,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall is a tool call the model produced.
type ToolCall struct {
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction is the name and JSON arguments of a tool call. Ollama sends
// arguments as a JSON object.
type ToolCallFunction struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// Tool declares a callable function.
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction is a function's name, description and JSON Schema parameters.
type ToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

// Options tunes generation. Temperature and TopP are pointers so an explicit
// zero is distinct from unset.
type Options struct {
	Temperature *float64 `json:"temperature,omitempty"`
	TopP        *float64 `json:"top_p,omitempty"`
	NumPredict  int      `json:"num_predict,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}

// ChatRequest is the native /api/chat request body.
type ChatRequest struct {
	Model     string          `json:"model"`
	Messages  []Message       `json:"messages"`
	Tools     []Tool          `json:"tools,omitempty"`
	Format    json.RawMessage `json:"format,omitempty"`
	Options   *Options        `json:"options,omitempty"`
	KeepAlive string          `json:"keep_alive,omitempty"`
	Stream    bool            `json:"stream"`
}

// ChatResponse is the native /api/chat response (and each streamed chunk).
type ChatResponse struct {
	Model           string  `json:"model"`
	CreatedAt       string  `json:"created_at"`
	Message         Message `json:"message"`
	Done            bool    `json:"done"`
	DoneReason      string  `json:"done_reason,omitempty"`
	PromptEvalCount int     `json:"prompt_eval_count"`
	EvalCount       int     `json:"eval_count"`
}

// ChatCompletion sends a native /api/chat request and returns the whole
// response. Use it for provider-specific options; use Generate for the shared,
// provider-agnostic path.
func (c *Client) ChatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	r := *req // do not mutate the caller's request
	r.Stream = false
	out, _, err := c.chatCompletion(ctx, &r)
	return out, err
}

func (c *Client) chatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, []byte, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, nil, err
	}
	data, status, err := c.send(ctx, http.MethodPost, "/api/chat", body)
	if err != nil {
		return nil, nil, err
	}
	if status != http.StatusOK {
		return nil, data, parseError(status, data)
	}
	var out ChatResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, data, err
	}
	return &out, data, nil
}

// Generate implements [ai.Client] over /api/chat.
func (c *Client) Generate(ctx context.Context, req *ai.Request) (*ai.Response, error) {
	cr, err := c.chatRequest(req, false)
	if err != nil {
		return nil, err
	}
	out, raw, err := c.chatCompletion(ctx, cr)
	if err != nil {
		return nil, err
	}
	resp := ollamaToResponse(out, raw)
	resp.Format = formatMode(req.Format)
	return resp, nil
}

// chatRequest converts an ai.Request into a native ChatRequest.
func (c *Client) chatRequest(req *ai.Request, stream bool) (*ChatRequest, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if err := checkHosted(req); err != nil {
		return nil, err
	}

	cr := &ChatRequest{
		Model:    req.Model,
		Messages: ollamaMessages(req),
		Stream:   stream,
	}
	for _, t := range req.Tools {
		schema := t.Schema
		if len(schema) == 0 {
			schema = json.RawMessage(`{"type":"object"}`)
		}
		cr.Tools = append(cr.Tools, Tool{
			Type: "function",
			Function: ToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  schema,
			},
		})
	}

	opt := &Options{Temperature: req.Temperature, TopP: req.TopP, Stop: req.Stop}
	if req.MaxTokens > 0 {
		opt.NumPredict = req.MaxTokens
	}
	if opt.Temperature != nil || opt.TopP != nil ||
		opt.NumPredict != 0 || len(opt.Stop) > 0 {
		cr.Options = opt
	}

	format, err := responseFormat(req.Format)
	if err != nil {
		return nil, err
	}
	cr.Format = format
	return cr, nil
}

// responseFormat renders an [ai.Format] as this provider's format field, which
// takes either the word "json" or a schema directly - there is no wrapper
// object around it. It returns nil when nothing was asked for.
func responseFormat(f *ai.Format) (json.RawMessage, error) {
	if f == nil || f.Type == ai.FormatText {
		return nil, nil
	}

	switch f.Type {
	case ai.FormatJSON:
		return json.RawMessage(`"json"`), nil
	case ai.FormatJSONSchema:
		return f.Schema, nil
	default:
		return nil, ai.ErrBadFormat
	}
}

// formatMode reports how the request's format was satisfied. Both shapes this
// driver accepts are enforced by the server itself, which constrains decoding
// rather than merely asking the model.
func formatMode(f *ai.Format) ai.FormatMode {
	if f == nil || f.Type == ai.FormatText {
		return ai.FormatNone
	}
	return ai.FormatNative
}

func ollamaMessages(req *ai.Request) []Message {
	var out []Message
	if req.System != "" {
		out = append(out, Message{Role: "system", Content: req.System})
	}
	for _, m := range req.Messages {
		switch m.Role {
		case ai.RoleSystem:
			out = append(out, Message{Role: "system", Content: partsText(m.Parts)})
		case ai.RoleTool:
			for _, p := range m.Parts {
				if tr, ok := p.(ai.ToolResult); ok {
					out = append(out, Message{Role: "tool", Content: tr.Content})
				}
			}
		case ai.RoleAssistant:
			msg := Message{Role: "assistant"}
			var text strings.Builder
			for _, p := range m.Parts {
				switch v := p.(type) {
				case ai.Text:
					text.WriteString(v.Text)
				case ai.ToolUse:
					msg.ToolCalls = append(msg.ToolCalls, ToolCall{
						Function: ToolCallFunction{Name: v.Name, Arguments: v.Input},
					})
				}
			}
			msg.Content = text.String()
			out = append(out, msg)
		default:
			msg := Message{Role: "user"}
			var text strings.Builder
			for _, p := range m.Parts {
				switch v := p.(type) {
				case ai.Text:
					text.WriteString(v.Text)
				case ai.Image:
					if len(v.Data) > 0 {
						msg.Images = append(msg.Images,
							base64.StdEncoding.EncodeToString(v.Data))
					}
				}
			}
			msg.Content = text.String()
			out = append(out, msg)
		}
	}
	return out
}

func partsText(parts []ai.Part) string {
	var b strings.Builder
	for _, p := range parts {
		if t, ok := p.(ai.Text); ok {
			b.WriteString(t.Text)
		}
	}
	return b.String()
}

func ollamaToResponse(cr *ChatResponse, raw []byte) *ai.Response {
	resp := &ai.Response{
		Model:      cr.Model,
		StopReason: cr.DoneReason,
		Usage: ai.Usage{
			InputTokens:  cr.PromptEvalCount,
			OutputTokens: cr.EvalCount,
		},
		Raw: append(json.RawMessage(nil), raw...),
	}
	if cr.Message.Content != "" {
		resp.Parts = append(resp.Parts, ai.Text{Text: cr.Message.Content})
	}
	for _, tc := range cr.Message.ToolCalls {
		resp.Parts = append(resp.Parts, ai.ToolUse{
			ID:    tc.Function.Name,
			Name:  tc.Function.Name,
			Input: tc.Function.Arguments,
		})
	}
	return resp
}
