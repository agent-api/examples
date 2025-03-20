package main

import (
	"context"
	"fmt"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/agent-api/core/agent"
	"github.com/agent-api/core/agent/bootstrap"
	"github.com/agent-api/ollama"
	"github.com/agent-api/ollama/models"
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

	// Create an Ollama provider
	opts := &ollama.ProviderOpts{
		Logger:  &logger,
		BaseURL: "http://localhost",
		Port:    11434,
	}

	provider := ollama.NewProvider(opts)
	provider.UseModel(ctx, models.GEMMA3_LATEST)

	// Create a new agent
	myAgent, err := agent.NewAgent(
		bootstrap.WithProvider(provider),
		bootstrap.WithLogger(&logger),
	)
	if err != nil {
		panic(err)
	}

	// Send a message to the agent
	response, err := myAgent.Run(
		ctx,
		agent.WithInput("Why is the sky blue?"),
	)
	if err != nil {
		logger.V(-1).Info("failed sending message to agent", "err", err)
		return
	}

	fmt.Println("Agent response:", response.Messages[1].Content)
}
