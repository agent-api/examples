package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/agent-api/anthropic"
	"github.com/agent-api/anthropic/models"
	"github.com/agent-api/core"
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

	// Create an Anthropic provider
	provider := anthropic.NewProvider(&anthropic.ProviderOpts{
		Logger: logger,
	})
	provider.UseModel(ctx, models.CLAUDE_3_5_SONNET)

	// Seed the message memory with the first user message
	memory := []*core.Message{
		{
			Role:    core.UserMessageRole,
			Content: "Why is the sky blue?",
		},
	}
	genOpts := &core.GenerateOptions{
		Messages: memory,
		Tools:    []*core.Tool{},
	}

	logger.Debug("sending message with generate options", "genOpts", genOpts)
	res, err := provider.Generate(ctx, genOpts)
	if err != nil {
		panic(err)
	}

	logger.Info("generate message finished", "result", res)
}
