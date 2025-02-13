package main

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	"github.com/agent-api/core/pkg/agent"
	"github.com/agent-api/core/types"
	"github.com/agent-api/ollama"
	"github.com/agent-api/ollama/models/qwen"
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

	// Create a zap logger
	var log logr.Logger
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}

	log = zapr.NewLogger(zapLog)

	// Create an Ollama provider
	opts := &ollama.ProviderOpts{
		BaseURL: "http://localhost",
		Port:    11434,
		Logger:  log,
	}
	provider := ollama.NewProvider(opts)
	provider.UseModel(ctx, qwen.QWEN2_5_LATEST)

	// Create a new agent
	agentConf := &agent.NewAgentConfig{
		Provider:     provider,
		SystemPrompt: "You are a helpful assistant.",
	}
	myAgent := agent.NewAgent(agentConf)

	// Register a simple calculator tool
	wrappedCalc, err := types.WrapToolFunction(calculator)
	if err != nil {
		log.Error(err, "could not wrap calculator function")
		return
	}

	err = myAgent.AddTool(types.Tool{
		Name:                "calculator",
		Description:         "Performs basic arithmetic operations: supported operations are 'add' and 'multiply'",
		WrappedToolFunction: wrappedCalc,
		JSONSchema:          []byte(jsonSchema),
	})
	if err != nil {
		log.Error(err, "adding agent tool unsuccessful")
		return
	}

	// Send a message to the agent
	response, err := myAgent.Run(ctx, "What is 5 + 3?", agent.DefaultStopCondition)
	if err != nil {
		log.Error(err, "failed sending message to agent")
		return
	}

	fmt.Println("Agent response:", response[len(response)-1].Message.Content)
}
