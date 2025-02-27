package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/agent-api/core/pkg/agent"
	"github.com/agent-api/core/types"
	"github.com/agent-api/ollama"

	"github.com/lmittmann/tint"
)

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
		Logger:  logger,
		BaseURL: "http://localhost",
		Port:    11434,
	}

	provider := ollama.NewProvider(opts)
	model := &types.Model{
		ID: "llama3.2-vision",
	}
	provider.UseModel(ctx, model)

	// Create a new agent
	agentConf := &agent.NewAgentConfig{
		Provider:     provider,
		Logger:       logger,
		SystemPrompt: "You are a professional image analyst.",
	}
	myAgent := agent.NewAgent(agentConf)

	// Send a message to the agent
	response := myAgent.Run(
		ctx,
		agent.WithInput("What is this image?"),
		agent.WithImagePath("/Users/jpmcb/Desktop/agent-api-devlog-000-1.png"),
	)
	if response.Err != nil {
		logger.Error("failed sending message to agent", "error", response.Err.Error())
		return
	}

	fmt.Println("Agent response:", response.Messages[1].Content)
}
