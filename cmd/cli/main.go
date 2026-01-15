package main

import (
	"fmt"
	"os"

	"github.com/andygeiss/cloud-native-utils/service"
	"github.com/andygeiss/memory-pipeline/internal/adapters/inbound"
	"github.com/andygeiss/memory-pipeline/internal/adapters/outbound"
	"github.com/andygeiss/memory-pipeline/internal/config"
	"github.com/andygeiss/memory-pipeline/internal/domain/extraction"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Println("Extraction completed successfully")
}

// printProgress displays the progress of a task in the console.
func printProgress(current, total int, desc string) {
	percent := float64(current) / float64(total) * 100
	fmt.Printf("\r%-20s: [%3.0f%%] %d/%d", desc, percent, current, total)
	if current == total {
		fmt.Println() // newline when done
	}
}

// run initializes and executes the memory extraction pipeline.
func run() error {
	// Create application context.
	ctx, cancel := service.Context()
	defer cancel()

	// Register shutdown hook.
	service.RegisterOnContextDone(ctx, func() {
		fmt.Println("Shutting down ...")
		os.Exit(0)
	})

	// Get configuration parameters.
	cfg := config.NewConfig()

	// Initialize inbound adapters.
	fs, err := inbound.NewFileWalker(cfg.MemorySourceDir, extraction.FilePath(cfg.MemoryStateFile), cfg.FileExtensions)
	if err != nil {
		return err
	}

	// Initialize outbound adapters.
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
			ProgressFn: printProgress,
		},
	)
	if err != nil {
		return err
	}

	// Run the extraction pipeline.
	return svc.Run()
}
