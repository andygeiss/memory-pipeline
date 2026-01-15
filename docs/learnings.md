# Learnings

General knowledge, facts, and concepts extracted from the codebase.

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/.github/agents/AGENTS-maintainer.md

The repository maintains a canonical AGENTS.md file to document the roles and interactions of all AI agents in the .github/agents directory.

---

Agents in .github/agents are responsible for maintaining various project documentation files like CONTEXT.md, README.md, and VENDOR.md.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/.github/agents/README-maintainer.md

A well-structured README.md includes a header, badges, one-line description, table of contents, overview, key features, architecture, installation, usage, testing, building/deployment, contributing, and license sections.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/.github/agents/VENDOR-maintainer.md

VENDOR.md serves as both a reference for current vendor usage and a governance tool to prevent reinvention of existing functionality.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/.github/agents/coding-assistant.md

The repository follows a structured architecture defined in CONTEXT.md, with clear conventions for directory layout, coding standards, and domain modeling patterns.

---

All development must align with the authoritative documents: CONTEXT.md (architecture and conventions), README.md (project purpose and usage), and VENDOR.md (approved vendor libraries).

---

Cross-cutting concerns should be implemented using approved vendor libraries specified in VENDOR.md instead of custom solutions to ensure reusability and maintainability.

---

The repository is designed as a reusable template, so all changes should preserve core architecture and directory layout while allowing for customization in domain-specific areas.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/AGENTS.md

The AGENTS.md file serves as a central index that describes all available AI agents, their purposes, and how they interact with one another.

---

The document hierarchy establishes a precedence order for resolving conflicts: CONTEXT.md for architecture, README.md for human-facing content, VENDOR.md for vendor usage, and agent files for agent-specific rules.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/CONTEXT.md

Memory Pipeline is a CLI tool that extracts structured knowledge notes from source files using a local LLM, generates embeddings, and produces searchable documentation.

---

The project follows Domain-Driven Design (DDD) with Hexagonal Architecture to separate business logic from infrastructure concerns.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/README.md

The memory pipeline generates semantic embeddings of extracted notes to enable powerful search and retrieval capabilities based on meaning rather than exact text matches.

---

The pipeline supports multiple input file types through configurable extensions and organizes extracted knowledge into four distinct categories: learning, pattern, cookbook, and decision.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/VENDOR.md

The cloud-native-utils library provides cross-cutting utilities for context management, security, and testing that should be preferred over custom implementations.

---

When adding new vendor dependencies, evaluate if existing tools like cloud-native-utils already provide the needed functionality.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/cmd/cli/main.go

The application supports configurable file extensions and memory state files, allowing it to process different types of input sources while maintaining a consistent internal representation.

---

The extraction pipeline includes progress reporting via a printProgress function that updates the console with percentage completion and current/total counts.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/cmd/cli/main_test.go

Benchmarking in Go uses the testing package with functions prefixed with 'Benchmark' and accepting a *testing.B parameter to control the benchmark loop.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/inbound/file_walker.go

The FileWalker uses a read-write mutex to synchronize access to its internal state map during concurrent operations like marking files as processed or retrieving pending files.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/inbound/file_walker_test.go

The file walker tracks file status changes through methods like MarkProcessing, MarkProcessed, and MarkError to maintain a consistent processing state.

---

File system timestamps may have limited granularity (e.g., 1 second on some systems), which can affect file change detection in tests or production scenarios.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/embedding_client.go

The embedding client handles API errors by checking the HTTP status code and parsing error messages from the response body, providing more context than generic network errors.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/embedding_client_test.go

The embedding client validates input parameters during initialization and returns specific errors for empty API key, base URL, or model.

---

The embedding client sends HTTP POST requests to the /embeddings endpoint with Authorization header and JSON payload containing the model and input text.

---

The embedding client preserves the original note kind (e.g., learning, pattern, cookbook) when processing and returning embedded notes.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/llm_client.go

The LLM client is designed to handle API errors gracefully by wrapping low-level errors with contextual ones and returning descriptive error messages.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/llm_client_test.go

The LLM client uses the 'Bearer' authentication scheme for API requests and sends content-type header as 'application/json'.

---

The LLM client sends a system message and a user message in the chat completion request, where the user message contains the content to be analyzed.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/markdown_writer_test.go

The markdown writer creates an index.md file that links to category-specific markdown files (learnings.md, patterns.md, cookbooks.md, decisions.md).

---

When multiple notes are written for the same file path, they are grouped under a single header for that file in the category markdown files.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/note_store.go

The NoteStore uses JSON marshaling with indentation for human-readable storage and ensures parent directories are created before writing files.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/note_store_test.go

The NoteStore requires a non-empty path; otherwise, it returns an error indicating the path is invalid.

---

The NoteStore loads existing notes from a file on instantiation, allowing for persistence and merging of new notes with existing ones.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/config/config.go

The configuration struct includes fields for memory management, OpenAI API settings, and file processing extensions, supporting a modular architecture for documentation and AI integration.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/domain/extraction/file.go

The extraction package defines a structured domain model for managing files and notes, with clear types and status constants to represent different states in the processing pipeline.

---

File status is tracked using a set of predefined constants: pending, processing, processed, and error, enabling clear state management during file handling.

---

Notes in the extraction system are categorized into four kinds: learning, pattern, cookbook, and decision, each serving a distinct purpose in knowledge organization.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/domain/extraction/ports.go

An EmbeddedNote is a MemoryNote that has been processed by an EmbeddingClient to include vector representations for semantic similarity searches.

---

The FileStore interface manages the lifecycle of files through states like pending, processing, and processed, with mechanisms to mark errors and track progress.

---

The LLMClient interface abstracts interactions with large language models for extracting structured memory notes from file contents.

---

The DocWriter interface is responsible for generating human-readable documentation from memory notes, with a finalize step to complete the output.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/domain/extraction/service.go

The extraction process involves multiple steps: fetching files, extracting notes via LLM, generating embeddings, saving notes, writing documentation, and updating file status.

---

The service ensures that all required dependencies are provided at construction time through a validation step in the configuration struct.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/domain/extraction/service_test.go

The extraction service treats each file as a unit of work, managing its state through a dedicated file store interface.

---

Memory notes are categorized into kinds such as learning, pattern, cookbook, and decision to support structured knowledge organization.

---

