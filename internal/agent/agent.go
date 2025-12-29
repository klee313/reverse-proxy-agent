// Package agent runs the supervisor loop that manages ssh lifecycle and restarts.
// It is invoked by cli and reports status via the ipc server.

package agent

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"reverse-proxy-agent/pkg/config"
	"reverse-proxy-agent/pkg/logging"
	"reverse-proxy-agent/pkg/monitor"
	"reverse-proxy-agent/pkg/state"
)

type Agent struct {
	cfg *config.Config
	sm  *state.StateMachine

	mu     sync.Mutex
	cmd    *exec.Cmd
	logger *logging.Logger

	forwardMu sync.Mutex

	startSuccessCount int
	startFailureCount int
	exitSuccessCount  int
	exitFailureCount  int
	lastTriggerReason string

	stopCh       chan struct{}
	stopOnce     sync.Once
	restartCount int
	lastExit     string

	policy  restartPolicy
	backoff *backoff

	errLines *lineBuffer

	lastSuccess time.Time
	lastClass   string
	lastTrigger time.Time
}

func New(cfg *config.Config) *Agent {
	return &Agent{
		cfg:     cfg,
		sm:      state.NewStateMachine(),
		stopCh:  make(chan struct{}),
		policy:  parseRestartPolicy(cfg.Agent.RestartPolicy),
		backoff: newBackoff(cfg.Agent.Restart),
	}
}

func (a *Agent) Start() error {
	if err := a.sm.Transition(state.StateConnecting); err != nil {
		return err
	}

	cmd, err := buildSSHCommand(a.cfg, a.currentRemoteForwards())
	if err != nil {
		_ = a.sm.Transition(state.StateStopped)
		a.recordStartFailure()
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = a.sm.Transition(state.StateStopped)
		a.recordStartFailure()
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		_ = a.sm.Transition(state.StateStopped)
		a.recordStartFailure()
		return err
	}

	if err := cmd.Start(); err != nil {
		_ = a.sm.Transition(state.StateStopped)
		a.recordStartFailure()
		return err
	}

	a.mu.Lock()
	a.cmd = cmd
	a.errLines = newLineBuffer(10)
	a.lastSuccess = time.Now()
	a.mu.Unlock()

	go drain(stdout, nil)
	go drain(stderr, a.errLines)

	if err := a.sm.Transition(state.StateConnected); err != nil {
		return err
	}
	a.recordStartSuccess()
	return nil
}

func (a *Agent) Stop() error {
	a.mu.Lock()
	cmd := a.cmd
	a.mu.Unlock()

	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Signal(os.Interrupt)
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()
		select {
		case <-time.After(3 * time.Second):
			_ = cmd.Process.Kill()
		case <-done:
		}
	}
	if err := a.sm.Transition(state.StateStopped); err != nil {
		return err
	}
	return nil
}

func (a *Agent) State() state.State {
	return a.sm.State()
}

func (a *Agent) ConfigSummary() string {
	return fmt.Sprintf("%s@%s:%d", a.cfg.SSH.User, a.cfg.SSH.Host, a.cfg.SSH.Port)
}

func (a *Agent) RunForeground() error {
	return fmt.Errorf("use RunWithLogger")
}

func (a *Agent) RunWithLogger(logger *logging.Logger) error {
	logger.Event("INFO", "agent_start", map[string]any{
		"summary": a.ConfigSummary(),
	})
	defer logger.Event("INFO", "agent_stop", nil)

	a.setLogger(logger)
	defer a.setLogger(nil)

	monitorCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var eventWG sync.WaitGroup
	eventWG.Add(1)
	go func() {
		defer eventWG.Done()
		monitor.StartSleepMonitor(monitorCtx, a.cfg.Agent, logger, func(reason string) {
			a.triggerRestart(logger, reason)
		})
	}()
	eventWG.Add(1)
	go func() {
		defer eventWG.Done()
		monitor.StartNetworkMonitor(monitorCtx, a.cfg.Agent, logger, func(reason string) {
			a.triggerRestart(logger, reason)
		})
	}()

	var periodicStop chan struct{}
	if a.cfg.Agent.PeriodicRestartSec > 0 {
		periodicStop = make(chan struct{})
		go a.periodicRestartLoop(logger, time.Duration(a.cfg.Agent.PeriodicRestartSec)*time.Second, periodicStop)
	}
	defer func() {
		cancel()
		eventWG.Wait()
		if periodicStop != nil {
			close(periodicStop)
		}
	}()

	go func() {
		<-a.stopCh
		cancel()
	}()

	for {
		select {
		case <-a.stopCh:
			logger.Info("stop requested")
			return a.Stop()
		default:
		}

		if err := a.Start(); err != nil {
			a.recordExit(fmt.Sprintf("start failed: %v", err))
			a.setLastTriggerReason("start failed")
			logger.Event("ERROR", "ssh_start_failed", map[string]any{
				"error": err.Error(),
			})
			a.mu.Lock()
			a.restartCount++
			a.mu.Unlock()
			if err := a.sleepWithBackoff(logger); err != nil {
				return err
			}
			continue
		}

		logger.Event("INFO", "ssh_started", map[string]any{
			"summary": a.ConfigSummary(),
		})
		a.mu.Lock()
		cmd := a.cmd
		a.mu.Unlock()
		if cmd == nil {
			a.recordExit("ssh command not started")
			logger.Error("ssh command not started")
			_ = a.sm.Transition(state.StateStopped)
			time.Sleep(2 * time.Second)
			continue
		}

		err := cmd.Wait()
		exitCode := 0
		if err != nil {
			a.recordExitFailure()
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = -1
			}
		} else {
			a.recordExitSuccess()
		}
		class := classifyExit(a.errLines, exitCode, err)
		a.setLastClass(class)
		exitMsg := formatExit(exitCode, err)
		if class != "clean" {
			exitMsg = fmt.Sprintf("%s (%s)", exitMsg, class)
		}
		a.recordExit(exitMsg)
		if err != nil {
			if summary := stderrSummary(a.errLines); summary != "" {
				logger.Event("ERROR", "ssh_exited", map[string]any{
					"exit":   exitMsg,
					"class":  class,
					"stderr": summary,
				})
			} else {
				logger.Event("ERROR", "ssh_exited", map[string]any{
					"exit":  exitMsg,
					"class": class,
				})
			}
		} else {
			logger.Event("INFO", "ssh_exited", map[string]any{
				"exit":  exitMsg,
				"class": class,
			})
		}

		_ = a.sm.Transition(state.StateStopped)

		a.mu.Lock()
		a.cmd = nil
		a.mu.Unlock()

		if !a.shouldRestart(exitCode, err, class) {
			logger.Info("restart policy: no restart (policy=%s, class=%s)", a.policyName(), class)
			return nil
		}
		if class == "auth" || class == "hostkey" {
			logger.Error("detected likely manual fix required; stopping auto-restart")
			return nil
		}
		if err == nil {
			a.backoff.Reset()
		}
		a.mu.Lock()
		a.restartCount++
		a.mu.Unlock()

		if err := a.sleepWithBackoff(logger); err != nil {
			return err
		}
	}
}

func (a *Agent) shouldRestart(exitCode int, err error, class string) bool {
	if class == "auth" || class == "hostkey" {
		return false
	}
	switch a.policy {
	case restartOnFailure:
		return err != nil || exitCode != 0
	default:
		return true
	}
}

func (a *Agent) policyName() string {
	if a.policy == restartOnFailure {
		return "on-failure"
	}
	return "always"
}

func (a *Agent) sleepWithBackoff(logger *logging.Logger) error {
	delay := a.backoff.Next()
	if delay <= 0 {
		return nil
	}
	logger.Info("restart scheduled in %s", delay.Round(time.Millisecond))
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-a.stopCh:
		logger.Info("stop requested during backoff")
		return a.Stop()
	case <-timer.C:
		return nil
	}
}

func (a *Agent) RequestStop() {
	a.stopOnce.Do(func() {
		close(a.stopCh)
		go func() {
			_ = a.Stop()
		}()
	})
}

func (a *Agent) RestartCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.restartCount
}

func (a *Agent) LastExitReason() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.lastExit
}

func (a *Agent) LastSuccess() time.Time {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.lastSuccess
}

func (a *Agent) LastClass() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.lastClass
}

func (a *Agent) LastTriggerReason() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.lastTriggerReason
}

func (a *Agent) StartSuccessCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.startSuccessCount
}

func (a *Agent) StartFailureCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.startFailureCount
}

func (a *Agent) ExitSuccessCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.exitSuccessCount
}

func (a *Agent) ExitFailureCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.exitFailureCount
}

func (a *Agent) CurrentBackoff() time.Duration {
	return a.backoff.Current()
}

func (a *Agent) recordExit(reason string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lastExit = reason
}

func (a *Agent) setLastClass(class string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lastClass = class
}

func (a *Agent) setLastTriggerReason(reason string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lastTriggerReason = reason
}

func (a *Agent) recordStartSuccess() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.startSuccessCount++
}

func (a *Agent) recordStartFailure() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.startFailureCount++
}

func (a *Agent) recordExitSuccess() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.exitSuccessCount++
}

func (a *Agent) recordExitFailure() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.exitFailureCount++
}

func (a *Agent) terminateProcess() {
	a.mu.Lock()
	cmd := a.cmd
	a.mu.Unlock()
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Signal(syscall.SIGTERM)
	}
}

func (a *Agent) periodicRestartLoop(logger *logging.Logger, interval time.Duration, stop <-chan struct{}) {
	if interval <= 0 {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-a.stopCh:
			return
		case <-ticker.C:
			if a.State() != state.StateConnected {
				continue
			}
			if !a.allowTrigger(time.Duration(a.cfg.Agent.Restart.DebounceMs) * time.Millisecond) {
				logger.Event("INFO", "restart_skipped", map[string]any{
					"reason": "periodic",
					"detail": "debounced",
				})
				continue
			}
			a.setLastTriggerReason("periodic")
			logger.Event("INFO", "restart_triggered", map[string]any{
				"reason": "periodic",
			})
			a.terminateProcess()
		}
	}
}

func (a *Agent) allowTrigger(window time.Duration) bool {
	if window <= 0 {
		return true
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	if !a.lastTrigger.IsZero() && now.Sub(a.lastTrigger) < window {
		return false
	}
	a.lastTrigger = now
	return true
}

func (a *Agent) triggerRestart(logger *logging.Logger, reason string) {
	a.requestRestart(reason, logger)
}

func (a *Agent) RequestRestart(reason string) {
	a.mu.Lock()
	logger := a.logger
	a.mu.Unlock()
	a.requestRestart(reason, logger)
}

func (a *Agent) requestRestart(reason string, logger *logging.Logger) {
	if a.State() != state.StateConnected {
		return
	}
	a.setLastTriggerReason(reason)
	if !a.allowTrigger(time.Duration(a.cfg.Agent.Restart.DebounceMs) * time.Millisecond) {
		if logger != nil {
			logger.Event("INFO", "restart_skipped", map[string]any{
				"reason": reason,
				"detail": "debounced",
			})
		}
		return
	}
	if logger != nil {
		logger.Event("INFO", "restart_triggered", map[string]any{
			"reason": reason,
		})
	}
	a.terminateProcess()
}

func (a *Agent) AddRemoteForward(forward string) (bool, error) {
	trimmed := strings.TrimSpace(forward)
	if trimmed == "" {
		return false, fmt.Errorf("remote forward is required")
	}
	a.forwardMu.Lock()
	defer a.forwardMu.Unlock()
	current := config.NormalizeRemoteForwards(a.cfg)
	for _, existing := range current {
		if existing == trimmed {
			return false, nil
		}
	}
	current = append(current, trimmed)
	config.SetRemoteForwards(a.cfg, current)
	a.RequestRestart("remote forward added")
	return true, nil
}

func (a *Agent) RemoveRemoteForward(forward string) (bool, error) {
	trimmed := strings.TrimSpace(forward)
	if trimmed == "" {
		return false, fmt.Errorf("remote forward is required")
	}
	a.forwardMu.Lock()
	defer a.forwardMu.Unlock()
	current := config.NormalizeRemoteForwards(a.cfg)
	next := make([]string, 0, len(current))
	removed := false
	for _, existing := range current {
		if existing == trimmed {
			removed = true
			continue
		}
		next = append(next, existing)
	}
	if !removed {
		return false, nil
	}
	if len(next) == 0 {
		return false, fmt.Errorf("at least one remote forward is required")
	}
	config.SetRemoteForwards(a.cfg, next)
	a.RequestRestart("remote forward removed")
	return true, nil
}

func (a *Agent) ClearRemoteForwards() bool {
	a.forwardMu.Lock()
	defer a.forwardMu.Unlock()
	current := config.NormalizeRemoteForwards(a.cfg)
	if len(current) == 0 {
		return false
	}
	config.SetRemoteForwards(a.cfg, nil)
	a.RequestStop()
	return true
}

func (a *Agent) currentRemoteForwards() []string {
	a.forwardMu.Lock()
	defer a.forwardMu.Unlock()
	return config.NormalizeRemoteForwards(a.cfg)
}

func (a *Agent) setLogger(logger *logging.Logger) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.logger = logger
}

func drain(r io.Reader, lines *lineBuffer) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if lines != nil {
			lines.Add(scanner.Text())
		}
	}
}

func stderrSummary(lines *lineBuffer) string {
	if lines == nil {
		return ""
	}
	list := lines.Lines()
	if len(list) == 0 {
		return ""
	}
	start := len(list) - 2
	if start < 0 {
		start = 0
	}
	summary := strings.Join(list[start:], " | ")
	if len(summary) > 200 {
		return summary[:200]
	}
	return summary
}

func classifyExit(lines *lineBuffer, exitCode int, err error) string {
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
