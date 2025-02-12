package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/agent-api/core/agent"
	"github.com/agent-api/core/pkg/defaultagent"
	"github.com/agent-api/core/tool"
	"github.com/agent-api/ollama-provider"
	"github.com/agent-api/ollama-provider/models/qwen"

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
	ctx := context.Background()

	// Create a standard library logger
	stdr.SetVerbosity(1)
	log := stdr.NewWithOptions(log.New(os.Stderr, "", log.LstdFlags), stdr.Options{
		LogCaller: stdr.All,
	})

	// Create an Ollama provider
	ollamaProviderOpts := &ollama.ProviderOpts{
		BaseURL: "http://localhost",
		Port:    11434,
		Logger:  log,
	}
	provider := ollama.NewProvider(ollamaProviderOpts)
	provider.UseModel(ctx, qwen.QWEN2_5)

	// Create a new agent
	agentConf := &agent.AgentConfig{
		Provider:     provider,
		SystemPrompt: "You are a helpful assistant.",
	}
	agent := defaultagent.NewAgent(agentConf)

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
	response, err := agent.Run(ctx, "What is 5 + 3?", defaultagent.DefaultStopCondition)
	if err != nil {
		log.Error(err, "failed sending message to agent")
		return
	}

	fmt.Println("Agent response:", response[len(response)-1].Message.Content)
}
