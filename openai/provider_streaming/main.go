package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/agent-api/core/types"
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
			logger.Error("stream timeout")
			panic("stream timeout")
		}
	}
}
