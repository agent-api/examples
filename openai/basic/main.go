package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/agent-api/core/pkg/agent"
	"github.com/agent-api/openai"
	"github.com/agent-api/openai/models/gpt4o"
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

	// Create an openai provider
	provider := openai.NewProvider(&openai.ProviderOpts{
		Logger: logger,
	})
	provider.UseModel(ctx, gpt4o.GPT4_O)

	// Create a new agent
	myAgent := agent.NewAgent(&agent.NewAgentConfig{
		Provider:     provider,
		Logger:       logger,
		SystemPrompt: "You are a helpful assistant.",
	})

	// Send a message to the agent
	response, err := myAgent.Run(ctx, "Why is the sky blue?", agent.DefaultStopCondition)
	if err != nil {
		logger.Error(err.Error(), "failed sending message to agent", err)
		return
	}

	fmt.Println("Agent response:", response[1].Message.Content)
}
