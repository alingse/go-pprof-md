package generator

import (
	"fmt"

	"github.com/alingse/go-pprof-md/internal/parser"
)

// getAIPrompt returns an AI-friendly analysis prompt based on profile type
func getAIPrompt(profileType parser.ProfileType) string {
	switch profileType {
	case parser.TypeCPU:
		return cpuAIPrompt()
	case parser.TypeHeap:
		return heapAIPrompt()
	case parser.TypeGoroutine:
		return goroutineAIPrompt()
	case parser.TypeMutex:
		return mutexAIPrompt()
	default:
		return genericAIPrompt()
	}
}

func cpuAIPrompt() string {
	return `
---

## AI Analysis Request

Please analyze this CPU profile and provide:

1. **Performance Bottlenecks**: Identify the top CPU-consuming functions and explain what they're doing.

2. **Call Chain Analysis**: For the top 5 functions by cumulative CPU usage:
   - Trace the call stack to identify the critical path
   - Explain which functions are called most frequently
   - Identify any unexpected or inefficient call patterns

3. **Optimization Recommendations**:
   - Which functions should be optimized first?
   - Are there any algorithmic improvements that could help?
   - Are there any caching opportunities?

4. **Code Investigation**: Suggest which source files and functions need detailed investigation.

Focus on actionable insights that can help improve the application's CPU performance.
`
}

func heapAIPrompt() string {
	return `
---

## AI Analysis Request

Please analyze this heap profile and provide:

1. **Memory Allocation Patterns**: Identify which functions are allocating the most memory.

2. **Memory Leak Investigation**:
   - Are there functions with high in-use memory that shouldn't be retaining objects?
   - Identify potential memory leaks by looking at long-lived allocations
   - Check for goroutine leaks or unclosed resources

3. **Allocation Hotspots**:
   - Which functions are allocating frequently but could be optimized?
   - Are there large object allocations that could be reduced?
   - Identify opportunities for object pooling or reuse

4. **Recommendations**:
   - What specific code changes would reduce memory usage?
   - Are there data structure optimizations possible?
   - Should we implement rate limiting or backpressure?

Focus on actionable insights to reduce memory usage and prevent leaks.
`
}

func goroutineAIPrompt() string {
	return `
---

## AI Analysis Request

Please analyze this goroutine profile and provide:

1. **Goroutine Distribution**: Identify where goroutines are created and blocked.

2. **Potential Issues**:
   - Are there goroutine leaks (goroutines that never exit)?
   - Identify goroutines stuck in blocking operations
   - Check for unlimited goroutine creation patterns

3. **Common Patterns**:
   - What are the main reasons goroutines are waiting?
   - Are there network I/O, channel operations, or mutex contention?
   - Identify any anti-patterns in goroutine usage

4. **Recommendations**:
   - How can we reduce the number of goroutines?
   - Should we implement worker pools instead of spawning goroutines?
   - Are there timeout or context cancellation issues?

Focus on actionable insights to optimize goroutine usage and prevent leaks.
`
}

func mutexAIPrompt() string {
	return `
---

## AI Analysis Request

Please analyze this mutex/lock contention profile and provide:

1. **Contention Hotspots**: Identify which locks have the most contention.

2. **Performance Impact**:
   - How much time is spent waiting for locks vs. doing actual work?
   - Which functions are most affected by lock contention?
   - Are there any deadlocks or long-held locks?

3. **Lock Usage Patterns**:
   - Are locks being held for too long?
   - Identify coarse-grained locks that could be split
   - Check for lock ordering issues that could cause deadlocks

4. **Recommendations**:
   - Which locks should be optimized first?
   - Can we use lock-free data structures instead?
   - Should we implement sharding or more fine-grained locking?
   - Are there opportunities to use sync/atomic or channels?

Focus on actionable insights to reduce lock contention and improve concurrent performance.
`
}

func genericAIPrompt() string {
	return `
---

## AI Analysis Request

Please analyze this profile and provide:

1. **Key Findings**: What are the most important patterns or issues in this profile?

2. **Resource Usage**: How are resources being consumed? Are there any inefficient patterns?

3. **Optimization Opportunities**: What specific changes would improve performance?

4. **Investigation Targets**: Which parts of the codebase need further investigation?

Focus on actionable insights.
`
}

// FormatBytes formats bytes into human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatDuration formats nanoseconds into human-readable format
func FormatDuration(nanos int64) string {
	ms := nanos / 1e6
	if ms < 1000 {
		return fmt.Sprintf("%d ms", ms)
	}
	s := ms / 1000
	if s < 60 {
		return fmt.Sprintf("%d s", s)
	}
	m := s / 60
	if m < 60 {
		return fmt.Sprintf("%d m %d s", m, s%60)
	}
	h := m / 60
	return fmt.Sprintf("%d h %d m", h, m%60)
}

// FormatNumber formats large numbers into human-readable format
func FormatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1_000_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	if n < 1_000_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	return fmt.Sprintf("%.1fG", float64(n)/1_000_000_000)
}
