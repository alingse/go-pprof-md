---
name: go-pprof-md
description: Use when you have a Go pprof binary file (.prof) and need to convert it to markdown format for AI analysis or human review
---

# go-pprof-md

## Overview

Convert Go pprof binary profiles to AI-readable markdown. Reads `.prof` files and outputs structured markdown with metrics, hotspots, and call stacks.

**Installation:**
```bash
go install github.com/alingse/go-pprof-md/cmd/go-pprof-md@latest
# or
git clone https://github.com/alingse/go-pprof-md.git && cd go-pprof-md && make build
```

## Commands

### show - Display a single profile

```bash
# Output markdown to stdout
go-pprof-md show <profile-file>

# Save to file
go-pprof-md show <profile-file> -o output.md
```

### diff - Compare two profiles

```bash
# Compare base vs new profile
go-pprof-md diff <base.prof> <new.prof>

# Save diff report to file
go-pprof-md diff base.prof new.prof -o regression.md
```

## Options

### show options

| Flag | Description | Default |
|------|-------------|---------|
| `-o, --output <file>` | Output file path | stdout |
| `-n, --top <number>` | Number of top functions to show | 20 |
| `-t, --type <type>` | Profile type: cpu, heap, goroutine, mutex | auto-detect |
| `--no-ai-prompt` | Disable AI analysis prompt section | false |

### diff options

| Flag | Description | Default |
|------|-------------|---------|
| `-o, --output <file>` | Output file path | stdout |
| `-n, --top <number>` | Number of top changed functions to show | 20 |
| `-b, --base-type <type>` | Base profile type | auto-detect |
| `-t, --new-type <type>` | New profile type | auto-detect |

## Examples

```bash
# Show CPU profile
go-pprof-md show cpu.prof

# Save heap analysis to file, show top 50 functions
go-pprof-md show heap.prof -o heap-report.md -n 50

# Compare profiles before and after optimization
go-pprof-md diff before.prof after.prof

# Regression test: compare against baseline
go-pprof-md diff baseline.prof current.prof -o regression.md

# Specify profile type explicitly
go-pprof-md show profile.prof -t goroutine -o report.md

# Disable AI prompt section
go-pprof-md show mutex.prof --no-ai-prompt
```

## Supported Profile Types

- **cpu**: CPU profiling samples
- **heap**: Memory allocation snapshots
- **goroutine**: Goroutine stack traces
- **mutex**: Mutex contention profiling

Auto-detection reads the profile file to determine type. For `diff`, both profiles must be the same type.
