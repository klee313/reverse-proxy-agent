// Package sshutil provides helpers for buffering and classifying SSH output.
// It is shared by agent and client supervisors.

package sshutil

import (
	"fmt"
	"strings"
	"sync"
)

type LineBuffer struct {
	mu    sync.Mutex
	limit int
	lines []string
}

func NewLineBuffer(limit int) *LineBuffer {
	return &LineBuffer{limit: limit}
}

func (b *LineBuffer) Add(line string) {
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

func (b *LineBuffer) Lines() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]string, len(b.lines))
	copy(out, b.lines)
	return out
}

func (b *LineBuffer) JoinedLower() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return strings.ToLower(strings.Join(b.lines, "\n"))
}

func ClassifyExit(lines *LineBuffer, exitCode int, err error) string {
	if err == nil && exitCode == 0 {
		return "clean"
	}
	if lines == nil {
		return "unknown"
	}

	text := lines.JoinedLower()
	switch {
	case strings.Contains(text, "permission denied"):
		return "auth"
	case strings.Contains(text, "host key verification failed"):
		return "hostkey"
	case strings.Contains(text, "could not resolve hostname"):
		return "dns"
	case strings.Contains(text, "name or service not known"):
		return "dns"
	case strings.Contains(text, "no route to host"):
		return "network"
	case strings.Contains(text, "connection refused"):
		return "refused"
	case strings.Contains(text, "operation timed out"):
		return "timeout"
	default:
		return "unknown"
	}
}

func FormatExit(exitCode int, err error) string {
	if err == nil {
		return "exit code 0"
	}
	return fmt.Sprintf("exit code %d (%v)", exitCode, err)
}
