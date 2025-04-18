package main

import (
	"context"

	"github.com/agent-api/core/agent"
	"github.com/agent-api/core/agent/bootstrap"
	"github.com/agent-api/googlegenai"
	"github.com/agent-api/openai"
	"github.com/agent-api/openai/models"
	"github.com/agent-api/pgvector"
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

	// Create a Google Gen AI provider
	embedder := googlegenai.NewEmbedder(&googlegenai.EmbedderOpts{
		Logger: &logger,
	})

	// Create an openai provider
	provider := openai.NewProvider(&openai.ProviderOpts{
		Logger: &logger,
	})
	provider.UseModel(ctx, models.GPT4_O)

	// making Pgvector connection
	pgv, err := pgvector.New(ctx, &pgvector.PgVectorStoreOpts{
		ConnectionString: "postgresql://admin:password@localhost:5432/test",
		TableName:        "test_table",
		Dimensions:       768,
		Embedder:         embedder,
	})
	if err != nil {
		panic(err)
	}

	_, err = pgv.Add(ctx, []string{"Fire Bolt - Cantrip - 120ft - You hurl a mote of fire at a creature or an object within range. Make a ranged spell attack against the target. On a hit, the target takes 1d10 Fire damage. A flammable object hit by this spell starts burning if it isn’t being worn or carried."})
	if err != nil {
		panic(err)
	}

	// Create a new agent
	myAgent, err := agent.NewAgent(
		bootstrap.WithProvider(provider),
		bootstrap.WithLogger(&logger),
		bootstrap.WithVectorStore(pgv),
	)
	if err != nil {
		panic(err)
	}

	myAgent.Run(
		ctx,
		agent.WithInput("How does the Fire Bolt spell work?"),
	)
}
