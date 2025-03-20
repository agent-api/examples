package main

import (
	"context"

	"github.com/agent-api/core"
	"github.com/agent-api/googlegenai"
	"github.com/agent-api/googlegenai/models"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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

	// Create a Google gen AI provider (Gemini)
	provider := googlegenai.NewProvider(&googlegenai.ProviderOpts{
		Logger: &logger,
	})
	provider.UseModel(ctx, models.GEMINI_1_5_FLASH)

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

	logger.Info("sending gen options", "opts", genOpts)
	res, err := provider.Generate(ctx, genOpts)
	if err != nil {
		panic(err)
	}

	logger.Info("generate message finished", "result", res)
}
