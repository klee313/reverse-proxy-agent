// Package client provides a Unix socket client for querying a running client.
// It is used by cli client commands such as status, logs, and metrics.

package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	"reverse-proxy-agent/pkg/config"
)

type Response struct {
	OK      bool              `json:"ok"`
	Message string            `json:"message,omitempty"`
	Data    map[string]string `json:"data,omitempty"`
	Logs    []string          `json:"logs,omitempty"`
}

func Query(cfg *config.Config, command string) (*Response, error) {
	socketPath, err := config.ClientSocketPath(cfg)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("connect to client: %w", err)
	}
	defer conn.Close()

	req := map[string]string{"command": command}
	enc := json.NewEncoder(conn)
	if err := enc.Encode(req); err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	reader := bufio.NewReader(conn)
	var resp Response
	if err := json.NewDecoder(reader).Decode(&resp); err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	return &resp, nil
}
