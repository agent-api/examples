package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/agent-api/core/types"
	"github.com/agent-api/gemini"
	"github.com/agent-api/gemini/models"
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

	// Create a gemini provider
	provider := gemini.NewProvider(&gemini.ProviderOpts{
		Logger: logger,
	})
	provider.UseModel(ctx, models.GEMINI_1_5_FLASH)

	// Seed the message memory with the first user message
	memory := []*types.Message{
		{
			Role:    types.UserMessageRole,
			Content: "Why is the sky blue?",
		},
	}
	genOpts := &types.GenerateOptions{
		Messages: memory,
		Tools:    []*types.Tool{},
	}

	logger.Debug("sending message with generate options", "genOpts", genOpts)
	res, err := provider.Generate(ctx, genOpts)
	if err != nil {
		panic(err)
	}

	logger.Info("generate message finished", "result", res)
}
