# Cookbooks

Step-by-step instructions and recipes for common tasks.

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/.github/agents/AGENTS-maintainer.md

To update AGENTS.md, scan .github/agents for all agent definition files, extract their purpose and ground-truth dependencies, then reconcile with the existing AGENTS.md content.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/.github/agents/README-maintainer.md

To update README.md, scan the repository for actual project structure, workflows, and configurations; cross-check against CONTEXT.md; and selectively update only affected sections.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/.github/agents/VENDOR-maintainer.md

To update VENDOR.md for a new or changed vendor: (1) review CONTEXT.md and README.md for alignment, (2) inspect actual codebase usage and dependency manifests, (3) research official vendor documentation, (4) draft or revise the VENDOR.md section with structured information, and (5) verify consistency with code and architecture.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/.github/agents/coding-assistant.md

To ensure code quality and consistency: (1) run `just fmt` after every code change to format Go source files, (2) run `just lint` to check for issues and resolve all reported problems, and (3) run `just test` to verify all tests pass.

---

When refactoring Go code, apply the following rules: (1) sort imports, constants, variables, and functions alphabetically, (2) order struct fields for memory alignment using betteralign, (3) maintain field literal assignments in struct definition order, and (4) sort switch cases alphabetically with default last.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/AGENTS.md

To add a new AI agent, create a markdown file in `.github/agents/<agent-name>.md` with purpose, ground-truth documents, workflow, and output rules, then run the AGENTS-maintainer to update the index.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/CONTEXT.md

To run the CLI locally: (1) ensure dependencies are installed with 'just setup', (2) execute 'just run' to start the tool, and (3) verify output in configured directories.

---

To build the project with PGO optimization: (1) run 'just profile' to generate a CPU profile, (2) build with 'go build -pgo=cpuprofile.pprof -o bin/cli ./cmd/cli'.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/README.md

To set up and run the memory pipeline locally: (1) start a local LLM server with an OpenAI-compatible API, (2) configure environment variables for source directory and file extensions, (3) run the CLI tool to process files, and (4) check generated JSON state and notes files along with browsable Markdown documentation.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/VENDOR.md

To initialize an application context with signal handling and shutdown hooks, use service.Context() and service.RegisterOnContextDone().

---

To safely parse environment variables with defaults, use security.ParseStringOrDefault(os.Getenv(key), defaultValue).

---

To perform content hashing with domain separation, use security.Hash(namespace, data) and encode the result as hex.

---

To write readable test assertions, use assert.That(t, message, actual, expected) from cloud-native-utils/assert.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/cmd/cli/main.go

To initialize and run the memory extraction pipeline, first create a context with service.Context(), register a shutdown handler, load configuration using config.NewConfig(), initialize inbound and outbound adapters, configure the extraction service with all required dependencies, and finally call svc.Run() to execute the pipeline.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/cmd/cli/main_test.go

To benchmark file system operations, create temporary directories with test files and use the testing.B.TempDir() method to ensure isolation and cleanup.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/inbound/file_walker.go

To initialize a FileWalker instance: (1) provide a source directory, (2) specify a state file path, and (3) define a list of valid file extensions; the adapter will load existing state if the file exists.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/inbound/file_walker_test.go

To initialize a file walker with validation, provide a non-empty source directory, a valid state file path, and a list of non-empty file extensions; otherwise, the constructor returns a specific error.

---

To read a file's content using the file walker, call ReadFile with a valid file path; it returns the file's content or an error if the file does not exist.

---

To process a file using the file walker, first call NextPending to get a pending file, then use MarkProcessing, MarkProcessed, or MarkError to update its status.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/embedding_client.go

To use the embedding client, initialize it with a valid API key, base URL, and model name. Then call the Embed method with a memory note to retrieve its embedding vector.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/embedding_client_test.go

To test an embedding client, create a mock HTTP server that simulates the expected API responses, then assert that the client correctly serializes requests and handles various response formats.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/llm_client.go

To interact with an LLM API, construct a request payload with system and user messages, set required headers including authorization and content type, send the POST request, and parse the JSON response.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/llm_client_test.go

To test an LLM client's behavior with a mock server, create an httptest.Server that simulates the API endpoint and verifies expected request parameters and headers.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/markdown_writer.go

To generate documentation using MarkdownWriter: (1) create a new instance with a valid output path, (2) call WriteDoc for each memory note to collect them, and (3) finalize the process by calling Finalize to write all notes to Markdown files.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/markdown_writer_test.go

To generate categorized markdown documentation from memory notes: (1) create a markdown writer with a target directory, (2) write individual notes using WriteDoc, (3) finalize the process with Finalize to create index and category files.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/note_store.go

To initialize a NoteStore: (1) provide a non-empty file path, (2) load existing notes from disk if they exist, (3) handle errors for missing or invalid files gracefully.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/adapters/outbound/note_store_test.go

To test the NoteStore, create a temporary directory using t.TempDir(), initialize the store with a path inside that directory, and verify file creation and content persistence through multiple SaveNote calls.

---

## /Users/andygeiss/Documents/github.com/loopforge-ai/memory-pipeline/internal/domain/extraction/service_test.go

To test the extraction service, mock implementations of all interfaces (FileStore, LLMClient, EmbeddingClient, NoteStore, DocWriter) should be used to isolate behavior and assert expected interactions.

---

