// Package monitor provides platform-specific sleep and network monitoring hooks.
// It is used by the agent and can be reused by client code.

package monitor

// Config captures the subset of settings needed by monitor implementations.
type Config struct {
	SleepCheckSec  int
	SleepGapSec    int
	NetworkPollSec int
}
