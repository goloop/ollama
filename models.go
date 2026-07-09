package ollama

import "context"

// Model describes a model installed on the server.
type Model struct {
	Name       string       `json:"name"`
	Model      string       `json:"model"`
	Size       int64        `json:"size"`
	ModifiedAt string       `json:"modified_at"`
	Details    ModelDetails `json:"details"`
}

// Models lists the models installed on the server (the /api/tags endpoint).
func (c *Client) Models(ctx context.Context) ([]Model, error) {
	var out struct {
		Models []Model `json:"models"`
	}
	if err := c.getJSON(ctx, "/api/tags", &out); err != nil {
		return nil, err
	}
	return out.Models, nil
}

// ModelInfo describes an installed model in detail (the /api/show endpoint).
// ModelInfo holds architecture-specific fields with dynamic keys, such as
// "llama.context_length"; Capabilities lists features like "completion" or
// "tools".
type ModelInfo struct {
	Modelfile    string         `json:"modelfile"`
	Parameters   string         `json:"parameters"`
	Template     string         `json:"template"`
	Details      ModelDetails   `json:"details"`
	ModelInfo    map[string]any `json:"model_info"`
	Capabilities []string       `json:"capabilities"`
}

// ModelDetails is the family and quantization summary shared by Model and
// ModelInfo.
type ModelDetails struct {
	Family            string `json:"family"`
	ParameterSize     string `json:"parameter_size"`
	QuantizationLevel string `json:"quantization_level"`
}

// Show returns detailed information about an installed model: its template,
// parameters, capabilities and architecture fields (the /api/show endpoint).
func (c *Client) Show(ctx context.Context, model string) (*ModelInfo, error) {
	var out ModelInfo
	if err := c.postJSON(ctx, "/api/show", map[string]string{"model": model}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
