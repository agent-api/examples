package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/agent-api/core/agent"
	"github.com/agent-api/core/pkg/defaultagent"
	"github.com/agent-api/core/tool"
	"github.com/agent-api/gsv"
	"github.com/agent-api/ollama-provider"
	"github.com/agent-api/ollama-provider/models/qwen"

	"github.com/go-logr/stdr"
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
	wrappedCalc := tool.WrapFunction(calculator)
	err = agent.AddTool(tool.Tool{
		Name:            "calculator",
		Description:     "Performs basic arithmetic operations: supported operations are 'add' and 'multiply'",
		WrappedFunction: wrappedCalc,
		JSONSchema:      schema,
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
