package ollama

import "context"

// Embed embeds one or more inputs with a model and returns their vectors in
// order, using the native /api/embed endpoint.
func (c *Client) Embed(
	ctx context.Context,
	model string,
	input ...string,
) ([][]float64, error) {
	body := struct {
		Model string   `json:"model"`
		Input []string `json:"input"`
	}{Model: model, Input: input}

	var out struct {
		Embeddings [][]float64 `json:"embeddings"`
	}
	if err := c.postJSON(ctx, "/api/embed", body, &out); err != nil {
		return nil, err
	}
	return out.Embeddings, nil
}
