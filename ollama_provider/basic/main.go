package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/agent-api/core/agent"
	"github.com/agent-api/core/pkg/defaultagent"
	"github.com/agent-api/ollama-provider"
	"github.com/agent-api/ollama-provider/models/qwen"

	"github.com/go-logr/stdr"
)

func main() {
	ctx := context.Background()

	// Create a standard library logger
	stdr.SetVerbosity(1)
	log := stdr.NewWithOptions(log.New(os.Stderr, "", log.LstdFlags), stdr.Options{
		LogCaller: stdr.All,
	})

	// Create an Ollama provider
	ollamaProviderOpts := &ollama.ProviderOpts{
		Logger:  log,
		BaseURL: "http://localhost",
		Port:    11434,
	}

	provider := ollama.NewProvider(ollamaProviderOpts)
	provider.UseModel(ctx, qwen.QWEN2_5)

	// Create a new agent
	agentConf := &agent.AgentConfig{
		Provider:     provider,
		SystemPrompt: "You are a helpful assistant.",
	}
	agent := defaultagent.NewAgent(agentConf)

	// Send a message to the agent
	response, err := agent.Run(ctx, "Why is the sky blue?", defaultagent.DefaultStopCondition)
	if err != nil {
		log.Error(err, "failed sending message to agent")
		return
	}

	fmt.Println("Agent response:", response[1].Message.Content)
}
