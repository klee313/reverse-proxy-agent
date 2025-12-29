//go:build !darwin

// Package monitor provides platform-specific sleep and network monitoring hooks.
// It is used by the agent and can be reused by client code.

package monitor

import (
	"context"
	"time"

	"reverse-proxy-agent/pkg/logging"
)

func StartSleepMonitor(ctx context.Context, cfg Config, logger *logging.Logger, onEvent func(reason string)) {
	if cfg.SleepCheckSec <= 0 {
		return
	}
	if onEvent == nil {
		onEvent = func(string) {}
	}
	logger.Info("sleep monitor: using gap-based fallback")
	sleepWatcher(ctx, logger, time.Duration(cfg.SleepCheckSec)*time.Second, time.Duration(cfg.SleepGapSec)*time.Second, onEvent)
}

func sleepWatcher(ctx context.Context, logger *logging.Logger, interval, gap time.Duration, onEvent func(reason string)) {
	if interval <= 0 {
		return
	}
	if gap <= 0 {
		gap = interval * 2
	}
	last := time.Now()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			if now.Sub(last) > gap {
				logger.Info("wake detected (gap=%s)", now.Sub(last).Truncate(time.Second))
				onEvent("wake")
			}
			last = now
		}
	}
}
