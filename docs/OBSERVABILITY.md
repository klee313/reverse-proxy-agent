# Observability

## Status

`rpa status` returns an `agent` section with:
- `state`: `STOPPED|CONNECTING|RUNNING`
- `summary`: `user@host:port`
- `remote_forwards`: comma-separated remote forward specs (optional)
- `uptime`: agent uptime
- `socket`: unix socket path
- `restarts`: restart count
- `last_exit`: last exit description
- `last_class`: exit classification
- `last_trigger`: last restart trigger reason
- `last_success_unix`: unix timestamp of the last SSH session that stayed up past the success grace period (optional)
- `tcp_check`: tcp reachability to the SSH host (`ok|failed|unknown`)
- `tcp_check_error`: tcp check error message (optional)
- `tcp_check_unix`: unix timestamp of the last tcp check (optional)
- `backoff_ms`: current backoff (optional)

`rpa status` returns a `client` section with:
- `state`: `STOPPED|CONNECTING|RUNNING`
- `summary`: `user@host:port (local=...)`
- `local_forwards`: comma-separated local forward specs (optional)
- `uptime`: client uptime
- `socket`: unix socket path
- `restarts`: restart count
- `last_exit`: last exit description
- `last_class`: exit classification
- `last_trigger`: last restart trigger reason
- `last_success_unix`: unix timestamp of the last SSH session that stayed up past the success grace period (optional)
- `tcp_check`: tcp reachability to the SSH host (`ok|failed|unknown`)
- `tcp_check_error`: tcp check error message (optional)
- `tcp_check_unix`: unix timestamp of the last tcp check (optional)
- `backoff_ms`: current backoff (optional)

### Metrics keys

`rpa metrics [agent]` returns:
- `rpa_agent_state`
- `rpa_agent_restart_total`
- `rpa_agent_uptime_sec`
- `rpa_agent_start_success_total`
- `rpa_agent_start_failure_total`
- `rpa_agent_exit_success_total`
- `rpa_agent_exit_failure_total`
- `rpa_agent_last_trigger`
- `rpa_agent_last_success_unix` (optional, set after the success grace period)
- `rpa_agent_backoff_ms` (optional)

`rpa metrics client` returns:
- `rpa_client_state`
- `rpa_client_restart_total`
- `rpa_client_uptime_sec`
- `rpa_client_start_success_total`
- `rpa_client_start_failure_total`
- `rpa_client_exit_success_total`
- `rpa_client_exit_failure_total`
- `rpa_client_last_trigger`
- `rpa_client_last_success_unix` (optional, set after the success grace period)
- `rpa_client_backoff_ms` (optional)
