# Architecture

This document describes key components of rpa and how recovery works.

## Components

`apps/rpa/internal/supervisor` drives process lifecycle, recovery policy, and monitoring.

- `apps/rpa/internal/supervisor`
  - Supervisor loop, restart policy, backoff handling.
- `apps/rpa/internal/agent` / `apps/rpa/internal/client`
  - Agent and client runtime entry points.
- `apps/rpa/pkg/monitor`
  - Sleep/network monitoring hooks.
- `apps/rpa/pkg/restart`
  - Backoff policy with exponential delay, jitter, and debounce window.
- `apps/rpa/pkg/sshutil`
  - Buffers SSH stderr and classifies exit failures for diagnostics.
- `apps/rpa/pkg/ipc`
  - Unix socket RPC for status/logs/metrics and runtime config changes.

## Recovery logic (detailed)

The recovery logic lives in `apps/rpa/internal/supervisor`. The flow is:

1) **Start attempt**
   - Build the SSH command (`apps/rpa/internal/agent/ssh.go` or `apps/rpa/internal/client/ssh.go`).
   - Transition state to CONNECTING, then RUNNING when the process starts.
   - Record start success/failure counters.

2) **Success marking (grace period)**
   - A "success" is recorded only after the SSH process stays alive for a short
     grace period (2 seconds). This avoids counting rapid failures as success.

3) **Monitor triggers**
   - Sleep/wake and network change monitors run in goroutines.
   - On events, `RequestRestart` is called with a debounce window to avoid
     restart storms (for example, multiple network events in quick succession).

4) **Process exit classification**
   - When SSH exits, stderr lines are buffered and classified into categories:
     `auth`, `hostkey`, `dns`, `network`, `refused`, `timeout`, `unknown`.
   - Classification is used for:
     - User-facing hints (`client run` and `doctor`).
     - Policy decisions (stop vs retry).

5) **Backoff and restart**
   - Backoff delay uses exponential policy with jitter.
   - Policy determines if restarts happen on all exits or only on failures.

## Key files

- CLI entry: `apps/rpa/cmd/rpa/main.go`
- CLI routing: `apps/rpa/internal/cli/cli.go`
- Agent runtime: `apps/rpa/internal/agent/agent.go`, `apps/rpa/internal/agent/ssh.go`
- Client runtime: `apps/rpa/internal/client/client.go`, `apps/rpa/internal/client/ssh.go`
- Supervisor core: `apps/rpa/internal/supervisor/supervisor.go`
- IPC servers: `apps/rpa/internal/agent/ipc/server.go`, `apps/rpa/internal/client/ipc/server.go`
