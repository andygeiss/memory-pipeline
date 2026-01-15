# Decisions

Architectural decisions, trade-offs, and rationale.

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/.github/agents/AGENTS-maintainer.md

When conflicts arise between documentation files, CONTEXT.md takes precedence for architecture and agent conventions, README.md for human-facing descriptions, and VENDOR.md for vendor-specific usage rules.

---

The AGENTS.md file should be kept at the repository root or under .github/AGENTS.md and must clearly list each agent's name, role summary, when to call it, and its source documents.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/.github/agents/README-maintainer.md

Only include badges in README.md for services or tools that are actually configured and functional within the repository, such as CI/CD, test coverage, language version, license, release, or package registry.

---

The README.md should avoid marketing fluff and focus on factual, clear descriptions of what the project does and how to use it, with all commands and examples being tested and verifiable.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/.github/agents/VENDOR-maintainer.md

The project adopts a vendor-first approach, preferring established third-party libraries over custom implementations to reduce duplication and improve maintainability.

---

The project prioritizes documenting vendor libraries that address cross-cutting concerns like testing, logging, validation, and concurrency, rather than domain-specific tools.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/.github/agents/coding-assistant.md

The team chose to enforce strict alphabetical ordering of imports, constants, variables, functions, and switch cases to improve code consistency and reduce merge conflicts.

---

Struct fields must be ordered for optimal memory alignment using the betteralign tool, not alphabetically, to reduce memory waste and improve performance.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/AGENTS.md

The team chose to define AI agents with specific responsibilities and ground-truth documents to ensure consistency, reduce ambiguity, and maintain high-quality documentation.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/CONTEXT.md

The team chose a local OpenAI-compatible LLM instead of a remote API to reduce latency, avoid external dependencies, and keep code private.

---

The project uses environment variables for configuration instead of config files to support containerized deployments and secure secret management.

---

The team adopted a hexagonal architecture to decouple business logic from external dependencies and support multiple implementations of adapters.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/README.md

The team chose to use local LLMs instead of remote APIs to ensure privacy, reduce latency, and eliminate reliance on external services for processing code.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/VENDOR.md

The project avoids external logging libraries like logrus or zap in favor of the standard library log package for simplicity.

---

The project uses cloud-native-utils/assert instead of testify for test assertions to maintain a consistent and minimal testing approach.

---

Configuration parsing is handled via cloud-native-utils/security.ParseStringOrDefault rather than viper or envconfig for simplicity and reduced dependencies.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/cmd/cli/main.go

The application uses a service-oriented design where external dependencies such as LLM clients and storage are injected into the core extraction service, enabling testability and flexibility in switching implementations.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/inbound/file_walker.go

The FileWalker uses JSON serialization to persist file processing state, with a structured fileState type that includes hash, path, reason, status, and modification time for each tracked file.

---

The FileWalker chooses to skip files without valid extensions and directories during directory traversal, reducing unnecessary processing and focusing only on relevant file types.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/inbound/file_walker_test.go

The system uses a state file to track the processing status of files, enabling resumable and idempotent processing by avoiding reprocessing already handled files.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/embedding_client.go

The team chose to use a custom HTTP client with a 30-second timeout to balance between responsiveness and avoiding indefinite hanging during API calls.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/embedding_client_test.go

The embedding client uses a fixed endpoint path '/embeddings' for all requests, which simplifies the client implementation and aligns with standard OpenAI API conventions.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/llm_client.go

The system uses a fixed system prompt to instruct the LLM on how to extract and format knowledge, ensuring consistency in the structure and quality of extracted notes.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/llm_client_test.go

The system chose to default unknown note kinds to 'learning' instead of failing, allowing for graceful handling of future or unexpected note types.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/markdown_writer.go

The team chose to organize documentation into four distinct categories—Learnings, Patterns, Cookbooks, and Decisions—to provide a clear taxonomy for different types of extracted knowledge.

---

The documentation system generates a main index.md file that links to category-specific Markdown files and includes summary statistics to give an overview of the knowledge base.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/markdown_writer_test.go

Empty documentation categories are filled with a placeholder message 'No notes in this category yet' to indicate that no content exists for that category.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/note_store.go

The team chose to store notes in a single JSON file rather than multiple files or a database, likely for simplicity and ease of version control during early development.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/note_store_test.go

The NoteStore uses JSON as the storage format to enable human-readable persistence and simple file-based operations, with support for embedding vectors stored alongside note metadata.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/config/config.go

The application uses environment variables for configuration management, with defaults chosen to support local development and testing without requiring explicit setup.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/domain/extraction/file.go

The extraction package separates file and note definitions into distinct sections to improve code organization and maintainability, allowing easier imports and usage across the application.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/domain/extraction/service.go

The service uses a sentinel error check to detect when no more pending files are available from the file store, allowing for clean loop termination.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/domain/extraction/service_test.go

The team chose to process files sequentially in a loop rather than in parallel, prioritizing simplicity and error isolation over performance.

---

The service uses a no-op progress function by default, allowing callers to optionally provide feedback without requiring changes to the core logic.

---

