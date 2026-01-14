# VENDOR.md

## Overview

This document describes the external vendor libraries used in Memory Pipeline and provides guidance on when and how to use them. The goal is to ensure consistency, encourage reuse of vendor functionality, and prevent duplicate implementations.

---

## Approved Vendor Libraries

### cloud-native-utils

- **Purpose:** Provides cloud-native utilities for Go applications including context management, security helpers, and test assertions.
- **Repository:** [github.com/andygeiss/cloud-native-utils](https://github.com/andygeiss/cloud-native-utils)
- **Version:** v0.4.12

#### Key Packages

| Package | Description |
|---------|-------------|
| `service` | Application lifecycle and context management |
| `security` | Cryptographic utilities and safe string parsing |
| `assert` | Test assertion helpers |

#### When to Use

- **Context Management:** Use `service.Context()` for creating cancellable application contexts with signal handling
- **Shutdown Hooks:** Use `service.RegisterOnContextDone()` for cleanup on graceful shutdown
- **Configuration Parsing:** Use `security.ParseStringOrDefault()` for safe environment variable parsing with defaults
- **Hashing:** Use `security.Hash()` for content hashing (e.g., file change detection)
- **Test Assertions:** Use `assert.That()` for readable test assertions

#### Integration Patterns

**Application Entry Point (`cmd/*/main.go`):**
```go
import "github.com/andygeiss/cloud-native-utils/service"

func run() error {
    ctx, cancel := service.Context()
    defer cancel()

    service.RegisterOnContextDone(ctx, func() {
        log.Println("shutting down...")
    })

    // ... application logic
}
```

**Configuration Loading (`internal/config/`):**
```go
import "github.com/andygeiss/cloud-native-utils/security"

func NewConfig() Config {
    return Config{
        Value: security.ParseStringOrDefault(os.Getenv("KEY"), "default"),
    }
}
```

**Content Hashing (`internal/adapters/`):**
```go
import "github.com/andygeiss/cloud-native-utils/security"

func computeHash(data []byte) string {
    hash := security.Hash("namespace", data)
    return hex.EncodeToString(hash)
}
```

**Test Assertions (`*_test.go`):**
```go
import "github.com/andygeiss/cloud-native-utils/assert"

func TestExample(t *testing.T) {
    result := doSomething()
    assert.That(t, "result must be true", result, true)
    assert.That(t, "err must be nil", err, nil)
}
```

#### Cautions

- The `security.Hash()` function requires a namespace string as the first argument for domain separation
- `service.Context()` listens for OS signals (SIGINT, SIGTERM) — do not create additional signal handlers
- Test assertions use a descriptive message as the second parameter for clear failure output

---

## Cross-Cutting Concerns

### Context and Lifecycle

**Always use:** `cloud-native-utils/service` for application context

Do not:
- Create raw `context.Background()` in main — use `service.Context()` instead
- Implement custom signal handling — the service package handles this

### Configuration

**Always use:** `cloud-native-utils/security` for parsing environment variables

Do not:
- Use raw `os.Getenv()` without default handling
- Implement custom "get or default" functions

### Testing

**Always use:** `cloud-native-utils/assert` for test assertions

Do not:
- Use raw `if` statements for test checks
- Import testify or other assertion libraries

### Hashing

**Always use:** `cloud-native-utils/security` for content hashing

Do not:
- Import `crypto/sha256` directly — use the vendor's `Hash()` function
- Create custom hashing utilities

---

## Indirect Dependencies

These libraries are transitive dependencies of `cloud-native-utils` and should **not** be imported directly:

| Library | Brought In By | Notes |
|---------|---------------|-------|
| `github.com/coreos/go-oidc/v3` | cloud-native-utils | OIDC support (not used in this project) |
| `github.com/go-jose/go-jose/v4` | go-oidc | JWT/JWE support |
| `golang.org/x/crypto` | cloud-native-utils | Cryptographic primitives |
| `golang.org/x/oauth2` | cloud-native-utils | OAuth2 client |
| `gopkg.in/yaml.v3` | cloud-native-utils | YAML parsing |

If you need YAML parsing or OAuth2 functionality, check if `cloud-native-utils` provides a wrapper before importing these directly.

---

## Vendors to Avoid

### testify

- **Why:** The project uses `cloud-native-utils/assert` for test assertions
- **Use instead:** `assert.That(t, message, actual, expected)`

### viper / envconfig

- **Why:** Configuration is simple; `security.ParseStringOrDefault()` suffices
- **Use instead:** Direct `os.Getenv()` with `security.ParseStringOrDefault()`

### logrus / zap

- **Why:** Standard library `log` is sufficient for this project's needs
- **Use instead:** `log.Println()`, `log.Fatalf()`

### Custom context utilities

- **Why:** `cloud-native-utils/service` provides complete lifecycle management
- **Use instead:** `service.Context()`, `service.RegisterOnContextDone()`

---

## Adding New Vendors

When adding a new vendor dependency:

1. **Check existing vendors first** — Can `cloud-native-utils` solve the problem?
2. **Document in this file** — Add a section with purpose, key packages, and usage patterns
3. **Update CONTEXT.md** — If the vendor affects architecture or conventions
4. **Prefer minimal dependencies** — Avoid large frameworks for small needs

### Evaluation Criteria

- Does it solve a problem not covered by existing vendors?
- Is it actively maintained?
- Does it have a permissive license (MIT, Apache 2.0, BSD)?
- Is the API stable and well-documented?
- Does it align with the project's architectural patterns?

---

## Version Management

Current vendor versions are pinned in `go.mod`. When upgrading:

1. Review the changelog for breaking changes
2. Run full test suite: `just test`
3. Update this document if APIs or patterns change
4. Commit `go.mod` and `go.sum` together
