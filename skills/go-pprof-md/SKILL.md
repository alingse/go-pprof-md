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

## Basic Usage

```bash
# Read pprof file and output markdown to stdout
go-pprof-md analyze <profile-file>

# Save to file
go-pprof-md analyze <profile-file> -o output.md
```

## Options

| Flag | Description | Default |
|------|-------------|---------|
| `-o, --output <file>` | Output file path | stdout |
| `-n, --top <number>` | Number of top functions to show | 20 |
| `-t, --type <type>` | Profile type: cpu, heap, goroutine, mutex | auto-detect |
| `--no-ai-prompt` | Disable AI analysis prompt section | false |

## Examples

```bash
# Convert CPU profile to markdown
go-pprof-md analyze cpu.prof

# Save heap analysis to file, show top 50 functions
go-pprof-md analyze heap.prof -o heap-report.md -n 50

# Specify profile type explicitly
go-pprof-md analyze profile.prof -t goroutine -o report.md

# Disable AI prompt section
go-pprof-md analyze mutex.prof --no-ai-prompt
```

## Supported Profile Types

- **cpu**: CPU profiling samples
- **heap**: Memory allocation snapshots
- **goroutine**: Goroutine stack traces
- **mutex**: Mutex contention profiling

Auto-detection reads the profile file to determine type.
