// Package logging provides a file-backed logger with an in-memory ring buffer for recent lines.
// It is used by the agent runtime and surfaced via IPC log queries.

package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"reverse-proxy-agent/pkg/config"
)

type LogBuffer struct {
	mu    sync.Mutex
	size  int
	lines []string
}

func newRingBuffer(size int) *LogBuffer {
	return &LogBuffer{size: size}
}

func NewLogBuffer() *LogBuffer {
	return newRingBuffer(200)
}

func (r *LogBuffer) Add(line string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.lines) >= r.size {
		copy(r.lines, r.lines[1:])
		r.lines[len(r.lines)-1] = line
		return
	}
	r.lines = append(r.lines, line)
}

func (r *LogBuffer) List() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.lines))
	copy(out, r.lines)
	return out
}

type Logger struct {
	path string
	ring *LogBuffer
	mu   sync.Mutex
}

func NewLogger(cfg *config.Config, ring *LogBuffer) (*Logger, error) {
	path, err := config.LogPath(cfg)
	if err != nil {
		return nil, err
	}
	return NewLoggerWithPath(path, ring)
}

func NewLoggerWithPath(path string, ring *LogBuffer) (*Logger, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}
	return &Logger{path: path, ring: ring}, nil
}

func (l *Logger) Info(format string, args ...any) {
	l.write("INFO", "message", fmt.Sprintf(format, args...), nil)
}

func (l *Logger) Error(format string, args ...any) {
	l.write("ERROR", "message", fmt.Sprintf(format, args...), nil)
}

func (l *Logger) Event(level, event string, fields map[string]any) {
	l.write(level, event, "", fields)
}

func (l *Logger) write(level, event, msg string, fields map[string]any) {
	entry := map[string]any{
		"ts":    time.Now().Format(time.RFC3339),
		"level": level,
		"event": event,
	}
	if msg != "" {
		entry["msg"] = msg
	}
	for k, v := range fields {
		entry[k] = v
	}
	encoded, err := json.Marshal(entry)
	if err != nil {
		encoded = []byte(fmt.Sprintf("{\"ts\":\"%s\",\"level\":\"%s\",\"event\":\"logger_error\",\"msg\":\"%s\"}",
			time.Now().Format(time.RFC3339), level, "json marshal failed"))
	}
	line := string(encoded)
	l.ring.Add(line)

	l.mu.Lock()
	defer l.mu.Unlock()
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(line + "\n")
}
