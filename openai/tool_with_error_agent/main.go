package main

import (
	"context"
	"fmt"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/agent-api/core"
	"github.com/agent-api/core/agent"
	"github.com/agent-api/core/agent/bootstrap"
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

var calls = 0

// calculator is a simple tool that can be used by an LLM
func calculator(ctx context.Context, args *calculatorParams) (interface{}, error) {
	if calls == 0 {
		calls++
		return nil, fmt.Errorf("internal error! PLEASE TRY AGAIN")
	}

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
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zLogger, err := config.Build()
	if err != nil {
		panic(err)
	}

	// Create a logr.Logger using zapr adapter
	logger := zapr.NewLogger(zLogger)

	// Create an OpenAI provider
	opts := &openai.ProviderOpts{
		Logger: &logger,
	}
	provider := openai.NewProvider(opts)
	provider.UseModel(ctx, models.GPT4_O)

	// Create a new agent
	myAgent, err := agent.NewAgent(
		bootstrap.WithProvider(provider),
		bootstrap.WithLogger(&logger),
	)
	if err != nil {
		panic(err)
	}

	// Register a simple calculator tool
	wrappedCalc, err := core.WrapToolFunction(calculator)
	if err != nil {
		logger.Error(err, "could not wrap calculator function", err)
		return
	}

	err = myAgent.AddTool(&core.Tool{
		Name:                "calculator",
		Description:         "Performs basic arithmetic operations: supported operations are 'add' and 'multiply'",
		WrappedToolFunction: wrappedCalc,
		JSONSchema:          []byte(jsonSchema),
	})
	if err != nil {
		logger.Error(err, "adding agent tool unsuccessful", err)
		return
	}

	// Send a message to the agent
	response, err := myAgent.Run(
		ctx,
		agent.WithInput("What is 987 * 123?"),
	)
	if err != nil {
		logger.Error(err, "failed sending message to agent", err)
		return
	}

	fmt.Println("Agent response:", response.Messages[len(response.Messages)-1].Content)
}
