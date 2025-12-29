// Package agent provides a line buffer used to classify recent stderr output.
// It is used by the agent to determine restart classes.

package agent

import (
	"strings"
	"sync"
)

type lineBuffer struct {
	mu    sync.Mutex
	limit int
	lines []string
}

func newLineBuffer(limit int) *lineBuffer {
	return &lineBuffer{limit: limit}
}

func (b *lineBuffer) Add(line string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.limit <= 0 {
		return
	}
	if len(b.lines) >= b.limit {
		copy(b.lines, b.lines[1:])
		b.lines[len(b.lines)-1] = line
		return
	}
	b.lines = append(b.lines, line)
}

func (b *lineBuffer) Lines() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]string, len(b.lines))
	copy(out, b.lines)
	return out
}

func (b *lineBuffer) JoinedLower() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return strings.ToLower(strings.Join(b.lines, "\n"))
}
