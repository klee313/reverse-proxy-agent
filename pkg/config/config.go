// Package config loads YAML config, applies defaults, and validates required fields.
// It is used by cli, agent, logging, and ipc to resolve runtime settings and paths.

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Agent   AgentConfig   `yaml:"agent"`
	SSH     SSHConfig     `yaml:"ssh"`
	Logging LoggingConfig `yaml:"logging"`
}

type AgentConfig struct {
	Name               string        `yaml:"name"`
	LaunchdLabel       string        `yaml:"launchd_label"`
	RestartPolicy      string        `yaml:"restart_policy"`
	Restart            RestartConfig `yaml:"restart"`
	PeriodicRestartSec int           `yaml:"periodic_restart_sec"`
	SleepCheckSec      int           `yaml:"sleep_check_sec"`
	SleepGapSec        int           `yaml:"sleep_gap_sec"`
	NetworkPollSec     int           `yaml:"network_poll_sec"`
}

type SSHConfig struct {
	User           string   `yaml:"user"`
	Host           string   `yaml:"host"`
	Port           int      `yaml:"port"`
	RemoteForward  string   `yaml:"remote_forward"`
	RemoteForwards []string `yaml:"remote_forwards"`
	IdentityFile   string   `yaml:"identity_file"`
	Options        []string `yaml:"options"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
}

type RestartConfig struct {
	MinDelayMs int     `yaml:"min_delay_ms"`
	MaxDelayMs int     `yaml:"max_delay_ms"`
	Factor     float64 `yaml:"factor"`
	Jitter     float64 `yaml:"jitter"`
	DebounceMs int     `yaml:"debounce_ms"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		return nil, errors.New("config path is empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	applyDefaults(&cfg)
	return &cfg, nil
}

func ApplyDefaults(cfg *Config) {
	if cfg == nil {
		return
	}
	applyDefaults(cfg)
}

func applyDefaults(cfg *Config) {
	if cfg.Agent.Name == "" {
		cfg.Agent.Name = "rpa-agent"
	}
	if cfg.Agent.LaunchdLabel == "" {
		cfg.Agent.LaunchdLabel = "com.rpa.agent"
	}
	if cfg.Agent.RestartPolicy == "" {
		cfg.Agent.RestartPolicy = "always"
	}
	if cfg.Agent.PeriodicRestartSec < 0 {
		cfg.Agent.PeriodicRestartSec = 0
	}
	if cfg.Agent.SleepCheckSec == 0 {
		cfg.Agent.SleepCheckSec = 5
	}
	if cfg.Agent.SleepGapSec == 0 {
		cfg.Agent.SleepGapSec = 30
	}
	if cfg.Agent.NetworkPollSec == 0 {
		cfg.Agent.NetworkPollSec = 5
	}
	if cfg.Agent.Restart.MinDelayMs == 0 {
		cfg.Agent.Restart.MinDelayMs = 2000
	}
	if cfg.Agent.Restart.MaxDelayMs == 0 {
		cfg.Agent.Restart.MaxDelayMs = 30000
	}
	if cfg.Agent.Restart.Factor == 0 {
		cfg.Agent.Restart.Factor = 2.0
	}
	if cfg.Agent.Restart.Jitter == 0 {
		cfg.Agent.Restart.Jitter = 0.2
	}
	if cfg.Agent.Restart.DebounceMs == 0 {
		cfg.Agent.Restart.DebounceMs = 2000
	}
	if cfg.SSH.Port == 0 {
		cfg.SSH.Port = 22
	}
	if len(cfg.SSH.Options) == 0 {
		cfg.SSH.Options = []string{
			"ServerAliveInterval=30",
			"ServerAliveCountMax=3",
		}
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Path == "" {
		cfg.Logging.Path = "~/.rpa/logs/agent.log"
	}
}

func Validate(cfg *Config) error {
	if cfg == nil {
		return errors.New("config is nil")
	}
	if strings.TrimSpace(cfg.SSH.Host) == "" {
		return errors.New("ssh.host is required")
	}
	if strings.TrimSpace(cfg.SSH.User) == "" {
		return errors.New("ssh.user is required")
	}
	if len(NormalizeRemoteForwards(cfg)) == 0 {
		return errors.New("ssh.remote_forward or ssh.remote_forwards is required")
	}
	if cfg.SSH.Port <= 0 {
		return fmt.Errorf("ssh.port must be > 0 (got %d)", cfg.SSH.Port)
	}
	switch strings.ToLower(cfg.Agent.RestartPolicy) {
	case "always", "on-failure":
	default:
		return fmt.Errorf("agent.restart_policy must be always or on-failure (got %q)", cfg.Agent.RestartPolicy)
	}
	if cfg.Agent.Restart.MinDelayMs < 0 || cfg.Agent.Restart.MaxDelayMs < 0 {
		return errors.New("agent.restart min/max delay must be >= 0")
	}
	if cfg.Agent.Restart.MaxDelayMs > 0 && cfg.Agent.Restart.MinDelayMs > cfg.Agent.Restart.MaxDelayMs {
		return errors.New("agent.restart min delay must be <= max delay")
	}
	if cfg.Agent.Restart.Factor < 1.0 {
		return errors.New("agent.restart factor must be >= 1.0")
	}
	if cfg.Agent.Restart.Jitter < 0 || cfg.Agent.Restart.Jitter > 1.0 {
		return errors.New("agent.restart jitter must be between 0 and 1")
	}
	if cfg.Agent.Restart.DebounceMs < 0 {
		return errors.New("agent.restart debounce_ms must be >= 0")
	}
	if cfg.Agent.PeriodicRestartSec < 0 {
		return errors.New("agent.periodic_restart_sec must be >= 0")
	}
	if cfg.Agent.SleepCheckSec < 0 {
		return errors.New("agent.sleep_check_sec must be >= 0")
	}
	if cfg.Agent.SleepGapSec < 0 {
		return errors.New("agent.sleep_gap_sec must be >= 0")
	}
	if cfg.Agent.NetworkPollSec < 0 {
		return errors.New("agent.network_poll_sec must be >= 0")
	}
	return nil
}

func NormalizeRemoteForwards(cfg *Config) []string {
	if cfg == nil {
		return nil
	}
	out := make([]string, 0, 1+len(cfg.SSH.RemoteForwards))
	seen := make(map[string]struct{})
	add := func(value string) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return
		}
		if _, ok := seen[trimmed]; ok {
			return
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	add(cfg.SSH.RemoteForward)
	for _, value := range cfg.SSH.RemoteForwards {
		add(value)
	}
	return out
}

func SetRemoteForwards(cfg *Config, forwards []string) {
	if cfg == nil {
		return
	}
	trimmed := make([]string, 0, len(forwards))
	seen := make(map[string]struct{})
	for _, value := range forwards {
		val := strings.TrimSpace(value)
		if val == "" {
			continue
		}
		if _, ok := seen[val]; ok {
			continue
		}
		seen[val] = struct{}{}
		trimmed = append(trimmed, val)
	}
	switch len(trimmed) {
	case 0:
		cfg.SSH.RemoteForward = ""
		cfg.SSH.RemoteForwards = nil
	case 1:
		cfg.SSH.RemoteForward = trimmed[0]
		cfg.SSH.RemoteForwards = nil
	default:
		cfg.SSH.RemoteForward = trimmed[0]
		cfg.SSH.RemoteForwards = append([]string(nil), trimmed[1:]...)
	}
}

func Save(path string, cfg *Config) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("config path is empty")
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create config dir: %w", err)
		}
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func SocketPath(cfg *Config) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	if cfg == nil {
		return "", errors.New("config is nil")
	}
	return filepath.Join(home, ".rpa", "agent.sock"), nil
}

func LogPath(cfg *Config) (string, error) {
	if cfg == nil {
		return "", errors.New("config is nil")
	}
	return expandHome(cfg.Logging.Path)
}

func expandHome(path string) (string, error) {
	if path == "" {
		return "", errors.New("path is empty")
	}
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	if path == "~" {
		return home, nil
	}
	return filepath.Join(home, strings.TrimPrefix(path, "~/")), nil
}
