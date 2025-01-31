package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/agent-api/core/agent"
	"github.com/agent-api/core/tool"
	"github.com/agent-api/gsv"
	"github.com/agent-api/ollama-provider"

	"github.com/go-logr/stdr"
)

type calculatorSchema struct {
	Operation *gsv.StringSchema `json:"operation"`
	A         *gsv.IntSchema    `json:"a"`
	B         *gsv.IntSchema    `json:"b"`
}

type calculatorParams struct {
	Operation string `json:"operation"`
	A         int    `json:"a"`
	B         int    `json:"b"`
}

func calculator(ctx context.Context, args *calculatorParams) (interface{}, error) {
	// Simple example implementation
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

	gsvSchema := &calculatorSchema{}
	gsvSchema.A = gsv.Int().Description("the first operand")
	gsvSchema.B = gsv.Int().Description("the first operand")
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
	ctx := context.Background()
	response, err := agent.SendMessage(ctx, "What is 5 + 3?")
	if err != nil {
		log.Error(err, "failed sending message to agent")
		return
	}

	fmt.Println("Agent response:", response.Content)
}
