package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"

	"github.com/agent-api/core/pkg/agent"
	"github.com/agent-api/core/types"
	"github.com/agent-api/openai"
	"github.com/agent-api/openai/models"
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

	// create a new std library logger
	logger := slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	)

	// Create an OpenAI provider
	opts := &openai.ProviderOpts{
		Logger: logger,
	}
	provider := openai.NewProvider(opts)
	provider.UseModel(ctx, models.GPT4_O)

	// Create a new agent
	myAgent := agent.NewAgent(&agent.NewAgentConfig{
		Provider:     provider,
		Logger:       logger,
		SystemPrompt: "You are a helpful assistant. YOU MUST ALWAYS USE AVAILABLE TOOLS.",
	})

	// Register a simple calculator tool
	wrappedCalc, err := types.WrapToolFunction(calculator)
	if err != nil {
		logger.Error(err.Error(), "could not wrap calculator function", err)
		return
	}

	err = myAgent.AddTool(types.Tool{
		Name:                "calculator",
		Description:         "Performs basic arithmetic operations: supported operations are 'add' and 'multiply'",
		WrappedToolFunction: wrappedCalc,
		JSONSchema:          []byte(jsonSchema),
	})
	if err != nil {
		logger.Error(err.Error(), "adding agent tool unsuccessful", err)
		return
	}

	// Send a message to the agent
	response := myAgent.Run(
		ctx,
		agent.WithInput("What is 987 * 123?"),
	)
	if response.Err != nil {
		logger.Error(response.Err.Error(), "failed sending message to agent", response.Err.Error())
		return
	}

	fmt.Println("Agent response:", response.Messages[len(response.Messages)-1].Content)
}
