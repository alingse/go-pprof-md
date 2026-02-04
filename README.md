# go-pprof-md

Convert Go pprof profiles to AI-readable markdown format.

## Overview

`go-pprof-md` is a CLI tool that converts Go pprof profiles (CPU, heap, goroutine, mutex) into markdown format optimized for AI analysis. The tool extracts key metrics, function hotspots, call stacks, and includes AI-optimized prompts to help you get actionable insights from your performance data.

## Features

- **Multiple Profile Types**: Supports CPU, heap, goroutine, and mutex profiles
- **Auto-Detection**: Automatically detects profile type from file content
- **AI-Optimized Output**: Includes structured prompts for AI analysis
- **Rich Statistics**: Shows summary statistics, top functions, and call stacks
- **Human-Readable**: Formats numbers, bytes, and durations in readable format

## Installation

```bash
go install github.com/alingse/go-pprof-md/cmd/go-pprof-md@latest
```

Or build from source:

```bash
git clone https://github.com/alingse/go-pprof-md.git
cd go-pprof-md
make build
```

## Usage

### Basic Usage

```bash
go-pprof-md analyze cpu.prof
```

This outputs the markdown to stdout. To save to a file:

```bash
go-pprof-md analyze cpu.prof -o analysis.md
```

### Options

- `-o, --output <file>`: Output file (default: stdout)
- `-n, --top <number>`: Number of top functions to display (default: 20)
- `-t, --type <type>`: Profile type: cpu, heap, goroutine, mutex (default: auto-detect)
- `--no-ai-prompt`: Disable AI analysis prompt

### Examples

```bash
# Analyze CPU profile and save to file
go-pprof-md analyze cpu.prof -o cpu-analysis.md

# Show top 50 functions
go-pprof-md analyze heap.prof -n 50

# Explicitly specify profile type
go-pprof-md analyze profile.prof -t goroutine

# Disable AI prompt
go-pprof-md analyze mutex.prof --no-ai-prompt
```

## Output Format

The generated markdown includes:

1. **Summary Statistics**: Profile-specific metrics (duration, samples, memory, etc.)
2. **Top Functions Table**: Ranked by resource consumption with percentages
3. **Call Stacks**: Detailed call chains for top functions
4. **AI Analysis Prompt**: Structured request for AI to provide insights

### Sample Output

```markdown
# cpu Profile Analysis

## Summary Statistics

- **Profile Duration:** 30 s
- **Total Samples:** 1234567
- **Sample Rate:** 100 Hz

## Top cpu Functions

| Rank | Function | File | CPU Samples | % of Total | Cumulative | Cumulative % |
|------|----------|------|-------------|------------|------------|--------------|
| 1 | `main.processData` | main.go:42 | 150000 | 12.15% | 150000 | 12.15% |
| 2 | `encoding/json.Unmarshal` | json/decode.go:123 | 100000 | 8.10% | 250000 | 20.25% |

### main.processData

**Call Stack:**
  → main.processData
    main.main
    runtime.main

---

## AI Analysis Request

Please analyze this CPU profile and provide:
...
```

## Creating pprof Files

### CPU Profile

```go
import (
    "os"
    "runtime/pprof"
)

func main() {
    f, _ := os.Create("cpu.prof")
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    // your code here
}
```

Or use `net/http/pprof`:

```bash
curl -o cpu.prof http://localhost:6060/debug/pprof/profile?seconds=30
```

### Heap Profile

```bash
curl -o heap.prof http://localhost:6060/debug/pprof/heap
```

### Goroutine Profile

```bash
curl -o goroutine.prof http://localhost:6060/debug/pprof/goroutine
```

### Mutex Profile

First enable mutex profiling:

```go
import "runtime"

func init() {
    runtime.SetMutexProfileFraction(1)
}
```

Then capture:

```bash
curl -o mutex.prof http://localhost:6060/debug/pprof/mutex
```

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
make build
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
