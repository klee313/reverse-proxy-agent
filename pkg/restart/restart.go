// Package restart defines restart policy parsing and backoff timing.
// It is shared by agent and client supervisors.

package restart

import (
	"math/rand"
	"strings"
	"time"

	"reverse-proxy-agent/pkg/config"
)

type Policy int

const (
	PolicyAlways Policy = iota
	PolicyOnFailure
)

func ParsePolicy(raw string) Policy {
	switch strings.ToLower(raw) {
	case "on-failure":
		return PolicyOnFailure
	default:
		return PolicyAlways
	}
}

func (p Policy) Name() string {
	if p == PolicyOnFailure {
		return "on-failure"
	}
	return "always"
}

type Backoff struct {
	min    time.Duration
	max    time.Duration
	factor float64
	jitter float64
	cur    time.Duration
	rng    *rand.Rand
}

func NewBackoff(cfg config.RestartConfig) *Backoff {
	return &Backoff{
		min:    time.Duration(cfg.MinDelayMs) * time.Millisecond,
		max:    time.Duration(cfg.MaxDelayMs) * time.Millisecond,
		factor: cfg.Factor,
		jitter: cfg.Jitter,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (b *Backoff) Next() time.Duration {
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

func (b *Backoff) Reset() {
	b.cur = 0
}

func (b *Backoff) ForceMax() {
	if b.max <= 0 {
		b.cur = b.min
		return
	}
	b.cur = b.max
}

func (b *Backoff) Current() time.Duration {
	return b.cur
}

func (b *Backoff) jittered(d time.Duration) time.Duration {
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
