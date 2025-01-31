package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/agent-api/core/agent"
	"github.com/agent-api/ollama-provider"

	"github.com/go-logr/stdr"
)

func main() {
	// Create a standard library logger
	stdr.SetVerbosity(1)
	log := stdr.NewWithOptions(log.New(os.Stderr, "", log.LstdFlags), stdr.Options{
		LogCaller: stdr.All,
	})

	// Create an Ollama provider
	provider := ollama.NewOllamaProvider(log, "http://localhost", 11434, "llama3.2")

	// Create a new agent
	agent := agent.NewAgent(provider)

	// Send a message to the agent
	ctx := context.Background()
	response, err := agent.SendMessage(ctx, "Why is the sky blue?")
	if err != nil {
		log.Error(err, "failed sending message to agent")
		return
	}

	fmt.Println("Agent response:", response.Content)
}
