package main

import (
	"context"

	"github.com/agent-api/core/agent"
	"github.com/agent-api/ollama"
	"github.com/agent-api/ollama/models"
	"github.com/agent-api/webscraper-agent"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const PROMPT string = "Please scrape https://johncodes.com/archive/2025/01-11-whats-an-ai-agent/ and summarize it."

func main() {
	ctx := context.Background()

	// Create a zap logger
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zLogger, err := config.Build()
	if err != nil {
		panic(err)
	}

	// Create a logr.Logger using zapr adapter
	logger := zapr.NewLogger(zLogger)

	// Create an Ollama provider
	opts := &ollama.ProviderOpts{
		Logger:  &logger,
		BaseURL: "http://localhost",
		Port:    11434,
	}

	provider := ollama.NewProvider(opts)
	provider.UseModel(ctx, models.QWEN2_5_LATEST)

	scraper, _ := webscraper.NewWebScraperAgent(&webscraper.WebScraperConfig{
		Provider: provider,
		Logger:   &logger,
		MaxSteps: 15,
	})

	result, err := scraper.Run(
		ctx,
		agent.WithInput(PROMPT),
	)
	if err != nil {
		panic(err)
	}

	logger.Info(result.Messages[len(result.Messages)-1].Content)
}
