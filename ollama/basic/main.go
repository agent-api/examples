package main

import (
	"context"
	"fmt"

	"github.com/agent-api/core/pkg/agent"
	"github.com/agent-api/ollama"
	"github.com/agent-api/ollama/models/qwen"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	// Create a zap logger
	var log logr.Logger
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}

	log = zapr.NewLogger(zapLog)

	// Create an Ollama provider
	opts := &ollama.ProviderOpts{
		Logger:  log,
		BaseURL: "http://localhost",
		Port:    11434,
	}

	provider := ollama.NewProvider(opts)
	provider.UseModel(ctx, qwen.QWEN2_5_LATEST)

	// Create a new agent
	agentConf := &agent.NewAgentConfig{
		Provider:     provider,
		SystemPrompt: "You are a helpful assistant.",
	}
	myAgent := agent.NewAgent(agentConf)

	// Send a message to the agent
	response, err := myAgent.Run(ctx, "Why is the sky blue?", agent.DefaultStopCondition)
	if err != nil {
		log.Error(err, "failed sending message to agent")
		return
	}

	fmt.Println("Agent response:", response[1].Message.Content)
}
