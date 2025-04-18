package main

import (
	"context"
	"fmt"

	"github.com/agent-api/core"
	"github.com/agent-api/googlegenai"
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

	logger.Info("making pgvec connection")
	pgv, err := pgvector.New(ctx, &pgvector.PgVectorStoreOpts{
		ConnectionString: "postgresql://admin:password@localhost:5432/test",
		TableName:        "test_table",
		Dimensions:       768,
		Embedder:         embedder,
	})
	if err != nil {
		panic(err)
	}

	logger.Info("making first vec")
	emb, err := pgv.Add(ctx, []string{"Hello world!"})
	if err != nil {
		panic(err)
	}

	fmt.Printf("first vec: %v\n", emb)

	logger.Info("searching")
	res, err := pgv.Search(ctx, &core.SearchParams{
		Query: "Howdy world!",
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Result: %v", res[0].Embedding.Content)
}
