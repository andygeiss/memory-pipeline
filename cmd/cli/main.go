package main

import (
	"log"

	"github.com/andygeiss/cloud-native-utils/service"
	"github.com/loopforge-ai/memory-pipeline/internal/adapters/inbound"
	"github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound"
	"github.com/loopforge-ai/memory-pipeline/internal/config"
	"github.com/loopforge-ai/memory-pipeline/internal/domain/extraction"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("main: %v", err)
	}
	log.Println("extraction completed successfully")
}

func run() error {
	// Create application context.
	ctx, cancel := service.Context()
	defer cancel()

	// Register shutdown hook.
	service.RegisterOnContextDone(ctx, func() {
		log.Println("main: shutting down memory-pipeline...")
	})

	// Get configuration parameters.
	cfg := config.NewConfig()

	// Initialize adapters.
	fs, err := inbound.NewFileWalker(cfg.MemorySourceDir, extraction.FilePath(cfg.MemoryStateFile), cfg.FileExtensions)
	if err != nil {
		return err
	}

	ec, err := outbound.NewEmbeddingClient(cfg.OpenAIAPIKey, cfg.OpenAIBaseURL, cfg.OpenAIEmbedModel)
	if err != nil {
		return err
	}

	llm, err := outbound.NewLLMClient(cfg.OpenAIAPIKey, cfg.OpenAIBaseURL, cfg.OpenAIChatModel)
	if err != nil {
		return err
	}

	ns, err := outbound.NewNoteStore(cfg.MemoryNotesFile)
	if err != nil {
		return err
	}

	// Create and configure the extraction service.
	svc, err := extraction.NewService(
		extraction.ServiceConfig{
			Embeddings: ec,
			Files:      fs,
			LLM:        llm,
			Notes:      ns,
		},
	)
	if err != nil {
		return err
	}

	// Run the extraction pipeline.
	return svc.Run()
}
