# Patterns

Reusable patterns, best practices, and conventions found in the code.

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/.github/agents/AGENTS-maintainer.md

Each agent maintains a specific domain of documentation and depends on authoritative source documents such as CONTEXT.md, README.md, and VENDOR.md for consistency.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/.github/agents/README-maintainer.md

The README.md file should be maintained as a human-first introduction to the project, focusing on accuracy, clarity, and usability for developers and contributors.

---

The README.md must align with CONTEXT.md in terms of architecture, conventions, and project rules, with CONTEXT.md being authoritative when there is a conflict.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/.github/agents/VENDOR-maintainer.md

The project maintains a dedicated VENDOR.md file to document all external libraries and dependencies, ensuring consistency with architectural boundaries defined in CONTEXT.md and positioned relative to the project's goals in README.md.

---

Vendor libraries are documented with structured sections including purpose, key packages, core capabilities, usage patterns, integration notes, and cautions to guide developers on when and how to use them.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/.github/agents/coding-assistant.md

Go files should be organized by functionality rather than by DDD element types like entities or value objects, to improve locality and readability following Go standard library idioms.

---

Domain packages should group related types, methods, and interfaces into cohesive files such as <aggregate>.go, service.go, and ports.go to maintain clear boundaries for adapters.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/AGENTS.md

The repository uses a set of specialized AI agents to maintain different aspects of the project, each with defined roles and ground-truth documents.

---

Each AI agent operates based on a set of ground-truth documents that define the project's architecture, conventions, and vendor usage patterns.

---

Agents collaborate in a defined workflow where code changes trigger updates to documentation, ensuring consistency across the project.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/CONTEXT.md

The codebase uses a ports and adapters pattern where domain logic defines interfaces (ports) and infrastructure implements them (adapters).

---

All configuration is managed via environment variables with defaults for local development, using a vendor library for safe parsing.

---

Error handling follows a pattern of domain-defined sentinel errors, wrapped with context using fmt.Errorf, and adapter-specific errors are kept private.

---

Testing is done with Go's standard testing package, using table-driven tests, temporary directories, and integration tags for specific test scenarios.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/README.md

The project follows hexagonal architecture with clear separation of domain logic from inbound and outbound adapters, enabling easy swapping of external dependencies like LLMs or storage systems.

---

File processing is stateful, using a JSON-based state file to track which files have been processed and avoid redundant work during subsequent runs.

---

Knowledge extraction from source files is done via structured prompts sent to a local LLM, producing categorized notes (learning, pattern, cookbook, decision) that are semantically embedded for search.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/VENDOR.md

Use cloud-native-utils/service for application context management and graceful shutdown handling, including signal listening and cleanup hooks.

---

Use cloud-native-utils/security for parsing environment variables with defaults and for content hashing with namespace separation.

---

Use cloud-native-utils/assert for test assertions to ensure readable failure messages and consistent testing practices.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/cmd/cli/main.go

The application follows a clean architecture pattern with clear separation of concerns between inbound adapters (file system access), outbound adapters (external services like LLMs and storage), and the core domain logic (extraction service).

---

The application uses a context-based shutdown mechanism with a cleanup hook registered via service.RegisterOnContextDone to ensure proper resource disposal when the application exits.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/cmd/cli/main_test.go

Use benchmark loops to repeatedly test code paths with clean state for each iteration, ensuring consistent and reliable performance measurements.

---

Mock external dependencies in benchmarks to isolate the performance of the system under test from external factors like network or disk I/O.

---

When benchmarking a service or pipeline, create mock implementations of all external interfaces to eliminate variability and focus on the core logic performance.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/inbound/file_walker.go

The FileWalker adapter uses a state file to persist the processing status of files, enabling resilient and restartable file processing pipelines.

---

The FileWalker implements a directory scanning pattern that tracks file modifications by comparing modification time and content hash to determine if files need reprocessing.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/inbound/file_walker_test.go

The file walker implements a stateful file processing workflow that tracks file statuses (pending, processing, processed, error) using a local state file to persist progress across multiple invocations.

---

File filtering in the walker is case-insensitive for file extensions, allowing matching of files with uppercase or lowercase extensions.

---

The file walker recursively scans directories to find files matching specified extensions, including nested subdirectories.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/embedding_client.go

The embedding client uses a structured approach to API communication with dedicated request and response types, enabling clear separation of concerns and easier testing.

---

The client implements input validation at both construction and usage time, ensuring required parameters like API key, base URL, and model are present before making requests.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/embedding_client_test.go

Use httptest.NewServer to mock HTTP endpoints in unit tests for external API clients, allowing isolation of the client logic from real network calls.

---

When testing HTTP clients, validate both successful responses and error cases including network errors, server errors, and invalid JSON responses.

---

Validate that API responses contain expected fields such as 'data' and 'embedding', and return specific errors when these fields are missing or malformed.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/llm_client.go

The LLM client follows a structured approach to extract knowledge from content by sending a system prompt and user content to an LLM API, then parsing the response into standardized memory note formats.

---

Memory notes are categorized into four kinds—learning, pattern, cookbook, and decision—with a default fallback to learning if the kind is unrecognized.

---

The LLM client validates required fields like API key, base URL, and model before initializing and ensures input content is not empty before initiating extraction.

---

Memory notes are generated with a unique ID using a security utility, and their content is mapped to a domain-specific type with associated metadata such as path and note kind.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/llm_client_test.go

The LLM client enforces validation of required configuration parameters (API key, base URL, model) during instantiation and returns specific error types for invalid inputs.

---

The LLM client uses a structured approach to communicate with external LLM APIs by sending POST requests to the /chat/completions endpoint with proper headers and JSON payload.

---

The LLM client expects structured JSON responses from the API and parses them to extract note data, defaulting unknown note kinds to 'learning' type.

---

The LLM client handles various error conditions from the API, including empty contents, invalid JSON responses, HTTP errors, and unauthorized access, by wrapping them in a consistent error type.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/markdown_writer.go

The MarkdownWriter uses a thread-safe map to collect and organize memory notes by kind, ensuring concurrent access is handled properly during documentation generation.

---

The documentation generation process follows a two-phase approach: first collecting all notes in memory, then writing them to disk in organized Markdown files with an index and category breakdown.

---

Memory notes are grouped by source file path within each category to maintain context and make it easier to trace the origin of extracted knowledge.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/markdown_writer_test.go

The markdown writer collects memory notes by category (learning, pattern, cookbook, decision) and groups them under file path headers when generating documentation files.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/note_store.go

The NoteStore implements a thread-safe in-memory cache with disk persistence using a read-write mutex to protect concurrent access.

---

The adapter pattern is used to persist embedded notes to disk as JSON, with a structured type (storedNote) that wraps the domain note data.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/note_store_test.go

The NoteStore adapter persists memory notes to a JSON file, supporting nested directory creation and maintaining note uniqueness by ID.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/config/config.go

Configuration values are loaded from environment variables with sensible defaults, using a helper function to safely parse strings or fall back to default values.

---

File extensions for processing are configurable via an environment variable, defaulting to a common set of text-based file types if not specified.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/domain/extraction/file.go

The system uses typed aliases (e.g., FileHash, FilePath) for domain identifiers to improve type safety and readability compared to using raw strings.

---

Constants are grouped by type and used with explicit naming to define valid states or categories, making the code more maintainable and self-documenting.

---

The MemoryNote structure includes metadata such as ID, content, kind, and path, which allows for structured storage and retrieval of knowledge items.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/domain/extraction/ports.go

The system uses a layered architecture with distinct interfaces for embedding, file I/O, LLM interaction, note storage, and documentation generation to enable loose coupling and testability.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/domain/extraction/service.go

The extraction service follows a sequential pipeline pattern for processing files, where each step must complete successfully before the next begins.

---

The service uses a configuration struct with validation to ensure all required dependencies are provided before creating the service instance.

---

Progress updates are reported using a dedicated progress function that accepts current position, total items, and a description.

---

The service handles file processing errors gracefully by marking files as errored and continuing with other files instead of failing the entire batch.

---

Each processing step in the pipeline is encapsulated in its own method, promoting separation of concerns and testability.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/domain/extraction/service_test.go

The extraction service follows a structured pipeline pattern: read files → extract notes with LLM → embed notes → save notes → update file status.

---

The service uses a configuration validation step to ensure all required dependencies are provided before starting execution.

---

Error handling in the extraction service is granular: individual file failures do not stop processing of other files, but overall errors are returned.

---

The service implements a progress callback to provide feedback during long-running operations, allowing for monitoring of processing state.

---

Each file is marked as processing, processed, or errored in the file store to track state and prevent duplicate work.

---

Memory notes are embedded with vector representations before being saved, enabling semantic search and similarity-based retrieval.

---

