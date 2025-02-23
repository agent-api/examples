package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/agent-api/core/pkg/agent"
	"github.com/agent-api/core/types"
	"github.com/agent-api/gsv"
	"github.com/agent-api/ollama"
	"github.com/agent-api/ollama/models"
	"github.com/lmittmann/tint"
)

type calculatorSchema struct {
	Operation *gsv.StringSchema `json:"operation"`
	A         *gsv.IntSchema    `json:"a"`
	B         *gsv.IntSchema    `json:"b"`
}

func calculator(ctx context.Context, args *calculatorSchema) (interface{}, error) {
	// Simple example implementation
	op, ok := args.Operation.Value()
	if !ok {
		return nil, fmt.Errorf("operation is not defined: %s", op)
	}

	a, ok := args.A.Value()
	if !ok {
		return nil, fmt.Errorf("a operand not defined: %s", op)
	}

	b, ok := args.B.Value()
	if !ok {
		return nil, fmt.Errorf("b operand not defined: %s", op)
	}

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

	// Create an Ollama provider
	opts := &ollama.ProviderOpts{
		BaseURL: "http://localhost",
		Port:    11434,
		Logger:  logger,
	}
	provider := ollama.NewProvider(opts)
	provider.UseModel(ctx, models.QWEN2_5_LATEST)

	// Create a new agent
	agentConf := &agent.NewAgentConfig{
		Provider:     provider,
		Logger:       logger,
		SystemPrompt: "You are a helpful assistant.",
	}
	myAgent := agent.NewAgent(agentConf)

	gsvSchema := &calculatorSchema{}
	gsvSchema.A = gsv.Int().Description("the first operand")
	gsvSchema.B = gsv.Int().Description("the second operand")
	gsvSchema.Operation = gsv.String().Description("The operation to perform. One of [add, multiply]")

	compileOpts := &gsv.CompileSchemaOpts{
		SchemaTitle:       "Calculator",
		SchemaDescription: "A simple calculator on ints",
	}

	schema, err := gsv.CompileSchema(gsvSchema, compileOpts)
	if err != nil {
		logger.Error(err.Error(), "could not compile schema", err)
		return
	}

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
		JSONSchema:          schema,
	})
	if err != nil {
		logger.Error(err.Error(), "adding agent tool unsuccessful", err)
		return
	}

	// Send a message to the agent
	response, err := myAgent.Run(ctx, "What is 5 + 3?", agent.DefaultStopCondition)
	if err != nil {
		logger.Error(err.Error(), "failed sending message to agent", err)
		return
	}

	fmt.Println("Agent response:", response[len(response)-1].Message.Content)
}
