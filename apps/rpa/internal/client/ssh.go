// Package client builds the ssh command line for local forwards.
// It is used by Client.Start when launching the tunnel.

package client

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"reverse-proxy-agent/pkg/config"
)

func buildSSHCommand(cfg *config.Config, localForwards []string) (*exec.Cmd, error) {
	if err := config.ValidateClient(cfg); err != nil {
		return nil, err
	}

	args := []string{
		"-N",
		"-T",
		"-o", "ExitOnForwardFailure=yes",
		"-o", "BatchMode=yes",
	}

	for _, forward := range localForwards {
		if strings.TrimSpace(forward) == "" {
			continue
		}
		args = append(args, "-L", forward)
	}

	if cfg.SSH.IdentityFile != "" {
		args = append(args, "-i", expandTilde(cfg.SSH.IdentityFile))
	}

	for _, opt := range cfg.SSH.Options {
		if strings.TrimSpace(opt) == "" {
			continue
		}
		args = append(args, "-o", opt)
	}

	if cfg.SSH.Port > 0 {
		args = append(args, "-p", strconv.Itoa(cfg.SSH.Port))
	}

	userHost := cfg.SSH.Host
	if cfg.SSH.User != "" {
		userHost = fmt.Sprintf("%s@%s", cfg.SSH.User, cfg.SSH.Host)
	}

	args = append(args, userHost)
	return exec.Command("ssh", args...), nil
}

func expandTilde(path string) string {
	if path == "" || path[0] != '~' {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	return filepath.Join(home, path[2:])
}
