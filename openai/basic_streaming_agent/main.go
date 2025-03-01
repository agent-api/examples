package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/agent-api/core/pkg/agent"
	"github.com/agent-api/openai"
	"github.com/agent-api/openai/models"
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
	provider.UseModel(ctx, models.GPT4_O)

	// Create a new agent
	myAgent := agent.NewAgent(&agent.NewAgentConfig{
		Provider:     provider,
		Logger:       logger,
		SystemPrompt: "You are a helpful assistant.",
	})

	result := myAgent.RunStream(
		ctx,
		agent.WithInput("Why is the sky blue?"),
	)

	for delta := range result.DeltaChan {
		print(delta)
	}
}
