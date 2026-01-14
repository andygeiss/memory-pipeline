# CONTEXT.md

## 1. Project Purpose

Memory Pipeline is a CLI tool that extracts structured knowledge notes from source files using a local LLM, generates embeddings, and stores them as a searchable knowledge base. It scans files with configurable extensions, processes them through an LLM to extract categorized notes (learnings, patterns, cookbooks, decisions), embeds them using a vector model, and persists them to JSON files.

This project follows Domain-Driven Design (DDD) with Hexagonal Architecture, serving as both a functional tool and a reference implementation for Go applications with clean separation between domain logic and infrastructure.

---

## 2. Technology Stack

- **Language:** Go 1.25+
- **External Library:** `github.com/andygeiss/cloud-native-utils` (context management, security utilities)
- **LLM Backend:** OpenAI-compatible API (designed for local LLMs like LM Studio)
- **Build Tools:** `just` (task runner), `podman`/`docker` (containerization), `golangci-lint` (linting)
- **Testing:** Go standard testing with coverage profiling

---

## 3. High-Level Architecture

The project uses **Hexagonal Architecture** (Ports and Adapters):

```
┌─────────────────────────────────────────────────────────────┐
│                         cmd/cli                             │
│                    (Application Entry)                      │
└─────────────────────────────┬───────────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────────┐
│                    internal/domain                          │
│              (Business Logic & Port Interfaces)             │
│                                                             │
│  extraction/                                                │
│    ├── ports.go      → Port interfaces (FileStore, LLM...)  │
│    ├── service.go    → Core extraction pipeline             │
│    └── file.go       → Domain types (File, Note, etc.)      │
└─────────────────────────────┬───────────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────────┐
│                   internal/adapters                         │
│                 (Port Implementations)                      │
│                                                             │
│  inbound/                                                   │
│    └── file_walker.go  → FileStore implementation           │
│                                                             │
│  outbound/                                                  │
│    ├── embedding_client.go  → EmbeddingClient impl          │
│    ├── llm_client.go        → LLMClient implementation      │
│    └── note_store.go        → NoteStore implementation      │
└─────────────────────────────────────────────────────────────┘
```

**Key Layers:**
- **Domain (`internal/domain/`):** Pure business logic with no external dependencies. Defines port interfaces that adapters must implement.
- **Adapters (`internal/adapters/`):** Implementations of domain ports. Split into `inbound/` (driving adapters) and `outbound/` (driven adapters).
- **Config (`internal/config/`):** Environment-based configuration loading.
- **Entry Point (`cmd/cli/`):** Wires adapters to domain services and runs the application.

---

## 4. Directory Structure (Contract)

```
memory-pipeline/
├── cmd/
│   └── cli/
│       ├── main.go           # Application entry point
│       └── main_test.go      # Benchmarks for PGO profiling
├── internal/
│   ├── adapters/
│   │   ├── inbound/          # Driving adapters (inputs)
│   │   │   ├── file_walker.go
│   │   │   └── file_walker_test.go
│   │   └── outbound/         # Driven adapters (outputs)
│   │       ├── embedding_client.go
│   │       ├── llm_client.go
│   │       └── note_store.go
│   ├── config/
│   │   └── config.go         # Environment configuration
│   └── domain/
│       └── extraction/       # Bounded context
│           ├── file.go       # Domain types and constants
│           ├── ports.go      # Port interfaces
│           ├── service.go    # Core service logic
│           └── service_test.go
├── bin/                      # Compiled binaries (gitignored)
├── .github/agents/           # AI agent definitions
├── .justfile                 # Task runner commands
├── .golangci.yml             # Linter configuration
├── Dockerfile                # Multi-stage container build
├── docker-compose.yml        # Service orchestration
└── go.mod                    # Go module definition
```

### Rules for New Code

1. **Domain logic** goes in `internal/domain/<bounded-context>/`
   - Define port interfaces in `ports.go`
   - Implement service orchestration in `service.go`
   - Define domain types in dedicated files (e.g., `file.go`)

2. **Adapters** go in `internal/adapters/`
   - `inbound/` for adapters that drive the application (file readers, HTTP handlers, etc.)
   - `outbound/` for adapters driven by the application (databases, external APIs, etc.)

3. **Tests** are co-located with their source files as `*_test.go`

4. **Configuration** lives in `internal/config/` and uses environment variables

5. **Entry points** live in `cmd/<app-name>/main.go`

---

## 5. Coding Conventions

### 5.1 General

- Small, focused packages with single responsibilities
- Prefer composition over inheritance
- Domain layer has zero external dependencies (except standard library)
- Adapters implement domain-defined interfaces
- Use constructor functions (`NewXxx`) for initialization with validation

### 5.2 Naming

| Element | Convention | Example |
|---------|------------|---------|
| Files | `snake_case.go` | `file_walker.go`, `llm_client.go` |
| Packages | `lowercase` | `extraction`, `inbound`, `outbound` |
| Types | `PascalCase` | `FileWalker`, `MemoryNote`, `NoteKind` |
| Interfaces | `PascalCase` + descriptive | `FileStore`, `EmbeddingClient` |
| Functions | `PascalCase` (exported), `camelCase` (private) | `NewService`, `parseNoteKind` |
| Constants | `PascalCase` | `FilePending`, `NoteLearning` |
| Errors | `Err` prefix + context | `ErrFileWalkerEmptySourceDir` |
| Domain types | Semantic wrappers | `FilePath`, `NoteContent`, `NodeID` |

### 5.3 Error Handling & Logging

- Define sentinel errors as package-level variables with descriptive prefixes
- Error messages follow pattern: `"<package>: <component> <description>"`
- Wrap errors with `fmt.Errorf("%w: %s", ErrXxx, detail)` for context
- Domain ports define sentinel errors that adapters must return (e.g., `ErrFileStoreNoMoreFiles`)
- Use `log.Fatalf` only in `main()` for fatal startup errors
- Use `log.Println` for informational messages

```go
// Domain-defined sentinel error (ports.go or file.go)
var ErrFileStoreNoMoreFiles = errors.New("extraction: file_store has no more pending files")

// Adapter-specific errors (adapter file)
var (
    ErrFileWalkerEmptySourceDir = errors.New("inbound: file_walker source_dir cannot be empty")
    ErrFileWalkerFileNotFound   = errors.New("inbound: file_walker file not found")
)
```

### 5.4 Testing

- Use Go's standard `testing` package
- Test files are co-located: `foo.go` → `foo_test.go`
- Use table-driven tests for multiple scenarios
- Use `t.Helper()` for test helper functions
- Test naming: `TestTypeName_MethodName_Scenario_ExpectedBehavior`
- Create temporary directories with `t.TempDir()` for file-based tests
- Integration tests use build tag `//go:build integration`

```go
func TestFileWalker_NextPending_EmptyDirectory_ReturnsError(t *testing.T) {
    // Arrange
    tmpDir := t.TempDir()
    // ...
    // Act
    _, err := fw.NextPending()
    // Assert
    assert.That(t, "err must not be nil", err != nil, true)
}
```

### 5.5 Formatting & Linting

- **Formatter:** `golangci-lint fmt ./...`
- **Linter:** `golangci-lint run ./...`
- Configuration in `.golangci.yml`
- Key disabled linters: `exhaustruct`, `ireturn`, `varnamelen`, `wrapcheck`
- Auto-rewrite `interface{}` → `any`

---

## 6. Cross-Cutting Concerns and Reusable Patterns

### Configuration

- All configuration via environment variables
- Defaults provided for local development (LM Studio on `localhost:1234`)
- Use `security.ParseStringOrDefault()` from vendor library for safe parsing

| Variable | Default | Purpose |
|----------|---------|---------|
| `MEMORY_SOURCE_DIR` | `.` | Directory to scan for files |
| `MEMORY_STATE_FILE` | `.memory-state.json` | Processing state persistence |
| `MEMORY_FILE` | `.memory-notes.json` | Extracted notes storage |
| `APP_FILE_EXTENSIONS` | `.md,.txt,.go` | File extensions to process |
| `OPENAI_BASE_URL` | `http://localhost:1234/v1` | LLM API endpoint |
| `OPENAI_CHAT_MODEL` | `qwen/qwen3-coder-30b` | Chat completion model |
| `OPENAI_EMBED_MODEL` | `text-embedding-qwen3-embedding-0.6b` | Embedding model |

### Context Management

- Use `service.Context()` from vendor library for graceful shutdown
- Register cleanup hooks with `service.RegisterOnContextDone()`

### Port/Adapter Pattern

When creating new adapters:
1. Implement the domain-defined interface exactly
2. Return domain-defined sentinel errors where specified
3. Keep adapter-specific errors private to the adapter package
4. Use constructor functions that validate required parameters

### File State Machine

Files progress through states: `pending` → `processing` → `processed` | `error`

---

## 7. Using This Repo as a Template

### Invariants (Must Preserve)

- Hexagonal architecture with domain/adapters separation
- Port interfaces defined in domain layer
- Environment-based configuration
- `just` commands for common operations
- `golangci-lint` configuration
- Multi-stage Docker build pattern

### Customization Points

- Add new bounded contexts in `internal/domain/<context>/`
- Add new adapters in `internal/adapters/inbound/` or `outbound/`
- Add new CLI commands in `cmd/<command>/`
- Extend configuration in `internal/config/config.go`
- Customize file extensions and LLM models via environment

### Steps to Create a New Project

1. Clone or copy this repository
2. Update `go.mod` module path
3. Update `APP_SHORTNAME` in `.env`
4. Modify domain types in `internal/domain/` for your use case
5. Implement adapters for your infrastructure needs
6. Update environment variables for your deployment

---

## 8. Key Commands & Workflows

```bash
# Install dependencies (macOS)
just setup

# Run the CLI locally
just run

# Run tests with coverage
just test

# Run integration tests (requires LM Studio)
just test-integration

# Run benchmarks and generate CPU profile for PGO
just profile

# Format code
just fmt

# Lint code
just lint

# Build Docker image
just build

# Build with PGO optimization
go build -pgo=cpuprofile.pprof -o bin/cli ./cmd/cli

# Start services
just up

# Stop services
just down
```

---

## 9. Important Notes & Constraints

### LLM Compatibility

- Designed for OpenAI-compatible APIs (LM Studio, Ollama, etc.)
- Does **not** use `response_format: json_object` (unsupported by many local LLMs)
- System prompt instructs JSON output format

### State Persistence

- `.memory-state.json` tracks file processing status
- `.memory-notes.json` stores extracted notes with embeddings
- Both files are local-only (typically gitignored)

### Performance Considerations

- Files are processed sequentially to respect LLM rate limits
- Hash-based change detection avoids reprocessing unchanged files
- PGO (Profile-Guided Optimization) supported via `just profile`
- Benchmarks in `cmd/cli/main_test.go` use `b.Loop()` (Go 1.24+) for accurate profiling

#### PGO Workflow

1. Run `just profile` to generate `cpuprofile.pprof` from benchmarks
2. Build with PGO: `go build -pgo=cpuprofile.pprof -o bin/cli ./cmd/cli`
3. The Dockerfile uses PGO by default if `cpuprofile.pprof` exists

### Security

- API keys loaded from environment variables
- File operations use restrictive permissions (0600, 0755)
- No secrets in code or default configurations

---

## 10. How AI Tools and RAG Should Use This File

This file is the **authoritative architectural reference** for this repository.

### For AI Agents

1. **Read `CONTEXT.md` first** before making significant changes
2. Follow the directory structure contract when adding new code
3. Use the established naming conventions and patterns
4. Return domain-defined sentinel errors from adapters
5. Place tests alongside source files

### For RAG Systems

- Index this file as high-priority context for repository-wide queries
- Use section 4 (Directory Structure) for file placement decisions
- Use section 5 (Coding Conventions) for style guidance
- Use section 6 (Cross-Cutting Concerns) for integration patterns

### Document Hierarchy

When conflicts arise:
- `CONTEXT.md` → Authoritative for architecture and conventions
- `README.md` → Authoritative for human-facing setup/usage
- `VENDOR.md` → Authoritative for vendor library patterns (if exists)
