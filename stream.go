package ollama

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"iter"
	"net/http"
	"strings"

	"github.com/goloop/ai"
)

// Stream implements [ai.Client] over /api/chat. Ollama streams newline-
// delimited JSON objects rather than Server-Sent Events.
func (c *Client) Stream(
	ctx context.Context,
	req *ai.Request,
) iter.Seq2[ai.Chunk, error] {
	return func(yield func(ai.Chunk, error) bool) {
		cr, err := c.chatRequest(req, true)
		if err != nil {
			yield(ai.Chunk{}, err)
			return
		}
		resp, err := c.openStream(ctx, cr)
		if err != nil {
			yield(ai.Chunk{}, err)
			return
		}
		defer resp.Body.Close()

		var usage ai.Usage
		for line, err := range jsonLines(resp.Body) {
			if err != nil {
				yield(ai.Chunk{}, err)
				return
			}
			var chunk ChatResponse
			if json.Unmarshal([]byte(line), &chunk) != nil {
				continue
			}
			if chunk.PromptEvalCount > 0 || chunk.EvalCount > 0 {
				usage = ai.Usage{
					InputTokens:  chunk.PromptEvalCount,
					OutputTokens: chunk.EvalCount,
				}
			}
			if chunk.Message.Content != "" {
				if !yield(ai.Chunk{Text: chunk.Message.Content, Raw: json.RawMessage(line)}, nil) {
					return
				}
			}
			for _, tc := range chunk.Message.ToolCalls {
				call := ai.ToolUse{
					ID:    tc.Function.Name,
					Name:  tc.Function.Name,
					Input: tc.Function.Arguments,
				}
				if !yield(ai.Chunk{ToolCall: &call, Raw: json.RawMessage(line)}, nil) {
					return
				}
			}
			if chunk.Done {
				break
			}
		}
		final := usage
		yield(ai.Chunk{Done: true, Usage: &final}, nil)
	}
}

// ChatStream sends a native streaming /api/chat request and yields each chunk
// as it arrives.
func (c *Client) ChatStream(
	ctx context.Context,
	req *ChatRequest,
) iter.Seq2[ChatResponse, error] {
	return func(yield func(ChatResponse, error) bool) {
		resp, err := c.openStream(ctx, req)
		if err != nil {
			yield(ChatResponse{}, err)
			return
		}
		defer resp.Body.Close()

		for line, err := range jsonLines(resp.Body) {
			if err != nil {
				yield(ChatResponse{}, err)
				return
			}
			var chunk ChatResponse
			if json.Unmarshal([]byte(line), &chunk) != nil {
				continue
			}
			if !yield(chunk, nil) {
				return
			}
		}
	}
}

// openStream opens the streaming /api/chat connection.
func (c *Client) openStream(ctx context.Context, req *ChatRequest) (*http.Response, error) {
	r := *req // do not mutate the caller's request
	r.Stream = true
	body, err := json.Marshal(&r)
	if err != nil {
		return nil, err
	}
	resp, err := c.opts.Do(ctx, http.MethodPost, c.opts.BaseURL+"/api/chat", body, c.headers())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, parseError(resp.StatusCode, data)
	}
	return resp, nil
}

// jsonLines iterates the non-empty newline-delimited lines of r.
func jsonLines(r io.Reader) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		sc := bufio.NewScanner(r)
		sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" {
				continue
			}
			if !yield(line, nil) {
				return
			}
		}
		if err := sc.Err(); err != nil {
			yield("", err)
		}
	}
}
