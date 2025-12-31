//go:build !darwin

// Package monitor provides platform-specific sleep and network monitoring hooks.
// It is used by the agent and can be reused by client code.

package monitor

import (
	"context"
	"time"

	"reverse-proxy-agent/pkg/logging"
)

func StartNetworkMonitor(ctx context.Context, cfg Config, logger *logging.Logger, onEvent func(reason string)) {
	if cfg.NetworkPollSec <= 0 {
		return
	}
	if onEvent == nil {
		onEvent = func(string) {}
	}
	logger.Info("network monitor: using polling fallback")
	networkWatcher(ctx, logger, time.Duration(cfg.NetworkPollSec)*time.Second, onEvent)
}
