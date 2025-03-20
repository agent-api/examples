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
	"github.com/agent-api/gsv"
	"github.com/agent-api/openai"
	"github.com/agent-api/openai/models"
)

type calculatorSchema struct {
	Operation *gsv.StringSchema `json:"operation"`
	A         *gsv.IntSchema    `json:"a"`
	B         *gsv.IntSchema    `json:"b"`
}

func calculator(ctx context.Context, args *calculatorSchema) (any, error) {
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
		logger.Error(err, "could not compile schema", err)
		return
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
		JSONSchema:          schema,
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
