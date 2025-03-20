package main

import (
	"context"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/agent-api/core/agent"
	"github.com/agent-api/core/agent/bootstrap"
	"github.com/agent-api/googlegenai"
	"github.com/agent-api/googlegenai/models"
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

	// Create a Google Gen AI provider
	provider := googlegenai.NewProvider(&googlegenai.ProviderOpts{
		Logger: &logger,
	})
	provider.UseModel(ctx, models.GEMINI_1_5_FLASH)

	// Create a new agent
	myAgent, err := agent.NewAgent(
		bootstrap.WithProvider(provider),
		bootstrap.WithLogger(&logger),
		bootstrap.WithSystemPrompt("You are a helpful assistant."),
	)
	if err != nil {
		panic(err)
	}

	result := myAgent.RunStream(
		ctx,
		agent.WithInput("What does Pomerium do?"),
	)

	for delta := range result.DeltaChan {
		print(delta)
	}
}
