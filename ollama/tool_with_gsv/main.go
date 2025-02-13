package main

import (
	"context"
	"fmt"

	"github.com/agent-api/core/pkg/agent"
	"github.com/agent-api/core/types"
	"github.com/agent-api/gsv"
	"github.com/agent-api/ollama"
	"github.com/agent-api/ollama/models/qwen"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
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
		log.Error(err, "could not compile schema")
		return
	}

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
		JSONSchema:          schema,
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
