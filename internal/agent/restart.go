// Package agent defines restart policy parsing and backoff timing.
// It is used by the agent loop for retry scheduling.

package agent

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"reverse-proxy-agent/pkg/config"
)

type restartPolicy int

const (
	restartAlways restartPolicy = iota
	restartOnFailure
)

func parseRestartPolicy(raw string) restartPolicy {
	switch strings.ToLower(raw) {
	case "on-failure":
		return restartOnFailure
	default:
		return restartAlways
	}
}

type backoff struct {
	min    time.Duration
	max    time.Duration
	factor float64
	jitter float64
	cur    time.Duration
	rng    *rand.Rand
}

func newBackoff(cfg config.RestartConfig) *backoff {
	return &backoff{
		min:    time.Duration(cfg.MinDelayMs) * time.Millisecond,
		max:    time.Duration(cfg.MaxDelayMs) * time.Millisecond,
		factor: cfg.Factor,
		jitter: cfg.Jitter,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (b *backoff) Next() time.Duration {
	if b.min <= 0 {
		return 0
	}
	if b.cur == 0 {
		b.cur = b.min
	} else {
		next := time.Duration(float64(b.cur) * b.factor)
		if b.max > 0 && next > b.max {
			next = b.max
		}
		b.cur = next
	}
	return b.jittered(b.cur)
}

func (b *backoff) Reset() {
	b.cur = 0
}

func (b *backoff) ForceMax() {
	if b.max <= 0 {
		b.cur = b.min
		return
	}
	b.cur = b.max
}

func (b *backoff) Current() time.Duration {
	return b.cur
}

func (b *backoff) jittered(d time.Duration) time.Duration {
	if b.jitter <= 0 {
		return d
	}
	delta := b.jitter * (b.rng.Float64()*2 - 1) // [-jitter, +jitter]
	out := time.Duration(float64(d) * (1 + delta))
	if out < 0 {
		return 0
	}
	return out
}

func formatExit(exitCode int, err error) string {
	if err == nil {
		return "exit code 0"
	}
	return fmt.Sprintf("exit code %d (%v)", exitCode, err)
}
