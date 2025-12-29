# Reverse Proxy Agent

Reverse Proxy Agent (rpa) is a CLI for managing a long-lived SSH remote forward
on macOS. It runs as a launchd service, monitors sleep/network changes, and
restarts the tunnel when needed.

## Features

- launchd integration (`agent up` / `agent down`)
- foreground mode for debugging (`agent run`)
- runtime status, logs, and metrics
- add/remove/clear remote forwards while running
- YAML config with defaults and validation

## Install

### From GitHub Releases (recommended)

Download the latest binary from Releases and place it on your PATH:

```sh
curl -L -o rpa https://github.com/<owner>/<repo>/releases/latest/download/rpa_<version>_darwin_arm64
chmod +x rpa
mv rpa /usr/local/bin/
```

### From source

```sh
go build -o rpa ./cmd/rpa
```

## Quick start

Initialize config (default: `~/.rpa/rpa.yaml`):

```sh
rpa init \
  --ssh-user ubuntu \
  --ssh-host example.com \
  --ssh-remote-forward "0.0.0.0:2222:localhost:22"
```

Start the agent (launchd):

```sh
rpa agent up
rpa status
```

Follow logs:

```sh
rpa logs --follow
```

Stop the agent:

```sh
rpa agent down
```

## Commands

```
rpa init --ssh-user user --ssh-host host --ssh-remote-forward spec [--config path]
rpa agent up --config path
rpa agent down --config path
rpa agent run --config path
rpa agent add --remote-forward spec --config path
rpa agent remove --remote-forward spec --config path
rpa agent clear --config path
rpa status --config path
rpa logs --config path [--follow]
rpa metrics --config path
```

## Configuration

Default config path: `~/.rpa/rpa.yaml` (override with `--config` or `RPA_CONFIG`).

Example:

```yaml
agent:
  name: "rpa-agent"
  launchd_label: "com.rpa.agent"
  restart_policy: "always"
  restart:
    min_delay_ms: 2000
    max_delay_ms: 30000
    factor: 2.0
    jitter: 0.2
    debounce_ms: 2000
  periodic_restart_sec: 3600
  sleep_check_sec: 5
  sleep_gap_sec: 30
  network_poll_sec: 5

ssh:
  user: "ubuntu"
  host: "example.com"
  port: 22
  remote_forward: "0.0.0.0:2222:localhost:22"
  remote_forwards:
    - "0.0.0.0:2223:localhost:23"
  identity_file: "~/.ssh/id_ed25519"
  options:
    - "ServerAliveInterval=30"
    - "ServerAliveCountMax=3"

logging:
  level: "info"
  path: "~/.rpa/logs/agent.log"
```

Notes:
- `ssh.remote_forward` and `ssh.remote_forwards` are merged and de-duplicated.
- `agent add/remove/clear` persists changes to the config file and applies them
  live when the agent is running.
- `agent clear` removes all forwards, stops the agent, and unloads launchd.
  To start again, run `rpa agent add ...` or `rpa init ...`.

## Observability

Logs are JSON lines and metrics are exposed via `rpa metrics`.
Schema details: `docs/OBSERVABILITY.md`.

## Development

Run from source:

```sh
go run ./cmd/rpa --help
```

## Releasing binaries (GitHub Releases)

This repo uses GoReleaser via GitHub Actions. Push a `v*` tag to publish binaries:

```sh
git tag v0.1.0
git push origin v0.1.0
```

Artifacts will appear under GitHub Releases for the tag.
