package main

import (
	"context"
	"fmt"

	"github.com/agent-api/core"
	"github.com/agent-api/core/agent"
	"github.com/agent-api/core/agent/bootstrap"
	"github.com/agent-api/gsv"
	"github.com/agent-api/ollama"
	"github.com/agent-api/ollama/models"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

	// Create a zap logger
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zLogger, err := config.Build()
	if err != nil {
		panic(err)
	}

	// Create a logr.Logger using zapr adapter
	logger := zapr.NewLogger(zLogger)

	// Create an Ollama provider
	opts := &ollama.ProviderOpts{
		BaseURL: "http://localhost",
		Port:    11434,
		Logger:  &logger,
	}
	provider := ollama.NewProvider(opts)
	provider.UseModel(ctx, models.QWEN2_5_LATEST)

	// Create a new agent
	myAgent, err := agent.NewAgent(
		bootstrap.WithProvider(provider),
		bootstrap.WithLogger(&logger),
		bootstrap.WithSystemPrompt("You are a professional image analyst."),
	)
	if err != nil {
		panic(err)
	}

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
		panic(err)
	}

	// Register a simple calculator tool
	wrappedCalc, err := core.WrapToolFunction(calculator)
	if err != nil {
		panic(err)
	}

	err = myAgent.AddTool(&core.Tool{
		Name:                "calculator",
		Description:         "Performs basic arithmetic operations: supported operations are 'add' and 'multiply'",
		WrappedToolFunction: wrappedCalc,
		JSONSchema:          schema,
	})
	if err != nil {
		panic(err)
	}

	// Send a message to the agent
	response, err := myAgent.Run(
		ctx,
		agent.WithInput("What is 5 + 3?"),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Agent response:", response.Messages[1].Content)
}
