# reverse-proxy-agent (rpa)

"내 홈서버를 외부에 노출하고 싶은데, 불특정 다수의 접속은 허용하지 않으면서
선별된 사용자에게만 SSH 기반 접근을 허용하고 싶다" 라는 목적을 위한
reverse proxy / ssh tunnel agent입니다.

## Why not `ngrok` / `tailscale`?

- ngrok, tailscale 역시 훌륭한 솔루션이지만, 홈서버 운영 환경에서 '불특정 다수에게 노출하지 않고'
  '특정 사용자에게만 접근을 허용'하는 설정을 아주 가볍게 하기는 쉽지 않습니다.
- 이 프로젝트는 SSH 기반으로 좁은 범위(선별된 사용자)의 접근만 허용하는 용도로 설계되었습니다.

## Use cases

- 홈서버 개발 테스트 환경에서 외부 접속을 열고 싶을 때
- 개인 서버를 특정 인원에게만 공개하고 싶을 때
- SSH 터널을 항상 살아있도록 관리하고 싶은 경우

## 설치

### GitHub Releases (권장)
```sh
curl -L -o rpa https://github.com/<owner>/<repo>/releases/latest/download/rpa_<version>_darwin_arm64
chmod +x rpa
mv rpa /usr/local/bin/
```

### 소스에서 빌드
```sh
cd apps/rpa
go build -o rpa ./cmd/rpa
```

## 빠른 시작

### Agent (원격 포워드)
```sh
rpa init \
  --ssh-user ubuntu \
  --ssh-host example.com \
  --remote-forward "0.0.0.0:2222:localhost:22"

rpa agent up
rpa status
rpa logs --follow
```

### Client (로컬 포워드)
```sh
rpa init \
  --ssh-user ubuntu \
  --ssh-host example.com \
  --local-forward "127.0.0.1:15432:127.0.0.1:5432"

rpa client up
rpa status
rpa logs --follow
```

## 구성 예시

```yaml
agent:
  name: "rpa-agent"
  launchd_label: "com.rpa.agent"
  restart_policy: "always"
  prevent_sleep: false
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

client:
  name: "rpa-client"
  launchd_label: "com.rpa.client"
  restart_policy: "always"
  prevent_sleep: false
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
  local_forwards:
    - "127.0.0.1:15432:127.0.0.1:5432"
    - "127.0.0.1:16379:127.0.0.1:6379"

ssh:
  user: "ubuntu"
  host: "example.com"
  port: 22
  check_sec: 5
  remote_forwards:
    - "0.0.0.0:2222:localhost:22"
    - "0.0.0.0:2223:localhost:23"
  identity_file: "~/.ssh/id_ed25519"
  options:
    - "ServerAliveInterval=30"
    - "ServerAliveCountMax=3"

logging:
  level: "info"
  path: "~/.rpa/logs/agent.log"

client_logging:
  level: "info"
  path: "~/.rpa/logs/client.log"
```

메모:
- `ssh.remote_forwards`는 중복 제거됩니다.
- 기본 SSH 옵션에 `ServerAlive*`와 `StrictHostKeyChecking=accept-new`가 포함됩니다(이미 지정한 경우 유지).
- `ssh.check_sec`은 SSH 호스트 TCP 체크 주기이며 `rpa status`에 표시됩니다.
- `agent clear`는 포워드를 모두 제거하고 서비스도 내려갑니다.

## 관측성

- 로그는 JSON 라인 형식
- `last_success_unix`는 연결이 2초 이상 유지된 뒤에만 기록됨
- status/metrics 상세 스키마: `docs/OBSERVABILITY.md`
- 구현/복구 로직 상세 설명: `docs/ARCHITECTURE.md`

## 릴리스 (GitHub Releases)

`v*` 태그를 푸시하면 GoReleaser가 바이너리를 업로드합니다.
```sh
git tag v0.1.0
git push origin v0.1.0
```

## 개발

```sh
cd apps/rpa
go test ./...
```

## Android 앱

Android 앱은 `apps/rpa-android`에 있습니다.

### 기술 스택
- Kotlin
- Jetpack Compose (UI)
- Material 3
- Foreground Service (상태바 알림 포함)
