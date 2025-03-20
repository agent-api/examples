package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/agent-api/core"
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

	logger.V(1).Info("sending message with generate options", "genOpts", genOpts)
	msgChan, deltaChan, errChan := provider.GenerateStream(ctx, genOpts)

	// Handle streamed messages and errors
	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				logger.Info("stream message channel closed")
				return
			}
			if msg != nil {
				logger.Info("received message",
					"role", msg.Role,
					"content", msg.Content,
					"tool_calls", msg.ToolCalls,
				)
			}

		case delta, ok := <-deltaChan:
			if !ok {
				logger.Info("stream delta chan closed")
				return
			}
			if delta != "" {
				print(delta)
			}

		case err, ok := <-errChan:
			if !ok {
				logger.Info("stream error chan closed")
				return
			}
			if err != nil {
				panic(err)
			}

		case <-time.After(30 * time.Second):
			logger.V(0).Error(fmt.Errorf("stream timeout"), "timeout")
			panic("stream timeout")
		}
	}
}
