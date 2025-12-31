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
