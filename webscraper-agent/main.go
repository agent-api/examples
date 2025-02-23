package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/agent-api/core/pkg/agent"
	"github.com/agent-api/ollama"
	"github.com/agent-api/ollama/models"
	"github.com/agent-api/webscraper-agent"
	"github.com/lmittmann/tint"
)

const PROMPT string = "Please scrape https://johncodes.com/archive/2025/01-11-whats-an-ai-agent/ and summarize it."

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
	provider.UseModel(ctx, models.QWEN2_5_LATEST)

	scraper, _ := webscraper.NewWebScraperAgent(&webscraper.WebScraperConfig{
		Provider: provider,
		Logger:   logger,
	})

	result, err := scraper.Run(ctx, PROMPT, agent.DefaultStopCondition)
	if err != nil {
		panic(err)
	}

	logger.Info(result[len(result)-1].Message.Content)
}
