package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/agent-api/core/agent"
	"github.com/agent-api/core/tool"
	"github.com/agent-api/ollama-provider"

	"github.com/go-logr/stdr"
)

const jsonSchema string = `{
  "title": "calculator",
  "description": "A simple calculator on ints",
  "type": "object",
  "properties": {
    "a": {
      "description": "The first operand",
      "type": "number"
    },
    "b": {
      "description": "The first operand",
      "type": "number"
    },
    "operation": {
      "description": "The operation to perform. One of [add, multiply]",
      "type": "string"
    }
  },
  "required": [
    "operation",
    "a",
    "b"
  ]
}`

type calculatorParams struct {
	Operation string `json:"operation"`
	A         int    `json:"a"`
	B         int    `json:"b"`
}

// calculator is a simple tool that can be used by an LLM
func calculator(ctx context.Context, args *calculatorParams) (interface{}, error) {
	println("Tool call!")
	op := args.Operation
	a := args.A
	b := args.B

	switch op {
	case "add":
		return a + b, nil
	case "multiply":
		return a * b, nil
	default:
		return nil, fmt.Errorf("unsupported operation: %s", op)
	}
}

func main() {
	// Create a standard library logger
	stdr.SetVerbosity(1) // Set the verbosity level
	log := stdr.NewWithOptions(log.New(os.Stderr, "", log.LstdFlags), stdr.Options{
		LogCaller: stdr.All, // Optional: log the calling function/file/line
	})

	// Create an Ollama provider
	provider := ollama.NewOllamaProvider(log, "http://localhost", 11434, "qwen2.5")

	// Create a new agent
	agent := agent.NewAgent(provider)

	// Register a simple calculator tool
	wrappedCalc := tool.WrapFunction(calculator)
	err := agent.AddTool(tool.Tool{
		Name:            "calculator",
		Description:     "Performs basic arithmetic operations: supported operations are 'add' and 'multiply'",
		WrappedFunction: wrappedCalc,
		JSONSchema:      []byte(jsonSchema),
	})

	if err != nil {
		log.Error(err, "adding agent tool unsuccessful")
		return
	}

	// Send a message to the agent
	ctx := context.Background()
	response, err := agent.SendMessage(ctx, "What is 5 + 3?")
	if err != nil {
		log.Error(err, "failed sending message to agent")
		return
	}

	fmt.Println("Agent response:", response.Content)
}
