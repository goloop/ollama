package ollama_test

import (
	"encoding/json"
	"fmt"

	"github.com/goloop/ai"
	"github.com/goloop/ollama"
)

func ExampleNew() {
	c := ollama.New("") // local server, no API key
	_ = c               // use c.Generate, c.Stream, c.ChatCompletion, ...
	fmt.Println(ollama.ModelLlama32)
	// Output: llama3.2
}

// ExampleClient_Generate builds a request. Sending it needs a running server,
// so this example only shows the shape.
func ExampleClient_Generate() {
	req := &ai.Request{
		Model: ollama.ModelLlama32,
		Messages: []ai.Message{
			ai.UserText("Name the capital of France."),
		},
	}
	fmt.Println(req.Model, len(req.Messages))
	// Output: llama3.2 1
}

// ExampleTool shows a tool definition passed with a request.
func ExampleTool() {
	tool := ai.Tool{
		Name:        "get_weather",
		Description: "Get the current weather for a city.",
		Schema:      json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}}}`),
	}
	fmt.Println(tool.Name)
	// Output: get_weather
}
