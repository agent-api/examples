package main

import (
	"context"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/agent-api/core/agent"
	"github.com/agent-api/core/agent/bootstrap"
	"github.com/agent-api/openai"
	"github.com/agent-api/openai/models"
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

	// Create an openai provider
	provider := openai.NewProvider(&openai.ProviderOpts{
		Logger: &logger,
	})
	provider.UseModel(ctx, models.GPT4_O)

	// Create a new agent
	myAgent, err := agent.NewAgent(
		bootstrap.WithProvider(provider),
		bootstrap.WithLogger(&logger),
	)
	if err != nil {
		panic(err)
	}

	result := myAgent.RunStream(
		ctx,
		agent.WithInput("Why is the sky blue?"),
	)

	for delta := range result.DeltaChan {
		print(delta)
	}
}
