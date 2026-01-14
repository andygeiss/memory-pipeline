# Memory Pipeline

[![Go Reference](https://pkg.go.dev/badge/github.com/andygeiss/memory-pipeline.svg)](https://pkg.go.dev/github.com/andygeiss/memory-pipeline)
[![License](https://img.shields.io/github/license/andygeiss/memory-pipeline)](https://github.com/andygeiss/memory-pipeline/blob/master/LICENSE)
[![Releases](https://img.shields.io/github/v/release/andygeiss/memory-pipeline)](https://github.com/andygeiss/memory-pipeline/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/andygeiss/memory-pipeline)](https://goreportcard.com/report/github.com/andygeiss/memory-pipeline)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/f9f01632dff14c448dbd4688abbd04e8)](https://app.codacy.com/gh/andygeiss/memory-pipeline/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)
[![Codacy Badge](https://app.codacy.com/project/badge/Coverage/f9f01632dff14c448dbd4688abbd04e8)](https://app.codacy.com/gh/andygeiss/memory-pipeline/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_coverage)

A CLI tool that extracts structured knowledge notes from source files using a local LLM, generates embeddings, and stores them as a searchable knowledge base.

## Overview

Memory Pipeline scans your codebase for files with configurable extensions (`.md`, `.txt`, `.go` by default), processes them through an LLM to extract categorized notes, generates vector embeddings, and persists everything to JSON files. It's designed to work with local LLMs like [LM Studio](https://lmstudio.ai/) via an OpenAI-compatible API.

### Note Categories

- **Learning** — General knowledge, facts, or concepts
- **Pattern** — Reusable patterns, best practices, or conventions
- **Cookbook** — Step-by-step instructions or recipes
- **Decision** — Architectural decisions, trade-offs, or rationale

## Features

- 🔍 **File Discovery** — Recursively scans directories for matching file extensions
- 🤖 **LLM Extraction** — Uses local LLMs to extract structured knowledge
- 📊 **Vector Embeddings** — Generates embeddings for semantic search
- 💾 **State Tracking** — Tracks processed files to avoid redundant work
- 🔄 **Change Detection** — Re-processes files when content changes

## Requirements

- Go 1.25 or later
- A local LLM server with OpenAI-compatible API (e.g., [LM Studio](https://lmstudio.ai/))
- [just](https://just.systems/) (optional, for task running)

## Installation

```bash
# Clone the repository
git clone https://github.com/andygeiss/memory-pipeline.git
cd memory-pipeline

# Install dependencies (macOS)
just setup

# Or manually install Go dependencies
go mod download
```

## Quick Start

1. **Start your local LLM server** (e.g., LM Studio on `localhost:1234`)

2. **Run the pipeline:**
   ```bash
   just run
   ```
   
   Or directly:
   ```bash
   go run ./cmd/cli/main.go
   ```

3. **Check the output:**
   - `.memory-state.json` — Processing state for each file
   - `.memory-notes.json` — Extracted notes with embeddings

## Commands

```bash
just run              # Run the CLI locally
just test             # Run tests with coverage
just test-integration # Run integration tests (requires LM Studio)
just profile          # Run benchmarks and generate CPU profile for PGO
just fmt              # Format code
just lint             # Lint code
just build            # Build Docker image
just up               # Start services
just down             # Stop services
just setup            # Install dependencies (macOS)
```

## Configuration

Configuration is done via environment variables. Create a `.env` file or export variables directly:

| Variable | Default | Description |
|----------|---------|-------------|
| `MEMORY_SOURCE_DIR` | `.` | Directory to scan for files |
| `MEMORY_STATE_FILE` | `.memory-state.json` | Processing state file |
| `MEMORY_FILE` | `.memory-notes.json` | Output notes file |
| `APP_FILE_EXTENSIONS` | `.md,.txt,.go` | Comma-separated file extensions |
| `OPENAI_BASE_URL` | `http://localhost:1234/v1` | LLM API endpoint |
| `OPENAI_API_KEY` | `not-used-in-local-llm-mode` | API key (if required) |
| `OPENAI_CHAT_MODEL` | `qwen/qwen3-coder-30b` | Chat model name |
| `OPENAI_EMBED_MODEL` | `text-embedding-qwen3-embedding-0.6b` | Embedding model name |

### Example

```bash
# Process only markdown files in ./docs
MEMORY_SOURCE_DIR=./docs APP_FILE_EXTENSIONS=.md just run
```

## Project Structure

```
memory-pipeline/
├── cmd/cli/              # Application entry point + benchmarks
├── internal/
│   ├── adapters/
│   │   ├── inbound/      # File walker (input adapter)
│   │   └── outbound/     # LLM, embedding, and storage adapters
│   ├── config/           # Environment configuration
│   └── domain/
│       └── extraction/   # Core business logic
├── .justfile             # Task runner commands
├── Dockerfile            # Container build
└── docker-compose.yml    # Service orchestration
```

The project follows **Hexagonal Architecture** (Ports and Adapters) with Domain-Driven Design principles. See [CONTEXT.md](CONTEXT.md) for detailed architectural documentation.

## How It Works

1. **Scan** — FileWalker discovers files matching configured extensions
2. **Track** — State manager tracks which files need processing
3. **Extract** — LLM analyzes file content and extracts structured notes
4. **Embed** — Embedding client generates vector representations
5. **Store** — Notes with embeddings are persisted to JSON

```
Files → FileWalker → LLMClient → EmbeddingClient → NoteStore
            ↓                                          ↓
      .memory-state.json                      .memory-notes.json
```

## Development

### Running Tests

```bash
# Unit tests with coverage
just test

# Integration tests (requires running LM Studio)
just test-integration
```

### Building

```bash
# Build binary
go build -o bin/cli ./cmd/cli

# Build with PGO optimization (after running just profile)
go build -pgo=cpuprofile.pprof -o bin/cli ./cmd/cli

# Build Docker image
just build
```

### Code Quality

```bash
# Format code
just fmt

# Run linter
just lint
```

## License

[MIT](LICENSE) © Andreas Geiß
