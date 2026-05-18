# godot-cli 설계서

> 버전: 1.0  
> 작성일: 2026-05-12  
> 상태: 설계 완료 (Planner → Architect → Critic 합의)  
> 레퍼런스: [godot-cli](https://github.com/Sqoff/godot-cli)

---

## 목차

1. [프로젝트 개요](#1-프로젝트-개요)
2. [아키텍처](#2-아키텍처)
3. [기술 스택](#3-기술-스택)
4. [기능 명세](#4-기능-명세)
5. [통신 프로토콜](#5-통신-프로토콜)
6. [인스턴스 관리](#6-인스턴스-관리)
7. [보안 모델](#7-보안-모델)
8. [디렉토리 구조](#8-디렉토리-구조)
9. [에러 처리](#9-에러-처리)
10. [구현 단계](#10-구현-단계)
11. [성공 기준](#11-성공-기준)
12. [위험 요소](#12-위험-요소)
13. [로드맵](#13-로드맵)
14. [의식적 절충 기록](#14-의식적-절충-기록)

---

## 1. 프로젝트 개요

### 목표

Godot Engine 에디터를 커맨드라인에서 직접 제어하는 CLI 도구. 개발자가 터미널에서 씬 실행, GDScript 실행, 상태 확인 등을 수행하여 Godot 개발 워크플로우를 자동화한다.

```bash
godot-cli status          # 연결된 에디터 정보 확인
godot-cli run             # 현재 씬 실행
godot-cli stop            # 씬 중지
godot-cli exec "print(Engine.get_version_info())"   # GDScript 실행
```

### 설계 원칙

| # | 원칙 | 설명 |
|---|------|------|
| 1 | **단일 바이너리, 제로 의존성** | `godot-cli` 바이너리 하나로 완결. 외부 런타임 불필요. |
| 2 | **Godot 표준 빌드 호환 최우선** | `.mono` 빌드 없이 표준 Godot 빌드에서 완전히 동작. |
| 3 | **점진적 발견** | 기본 커맨드(`run`, `status`)는 즉시 사용, 고급 기능(`exec`)은 필요 시 발견. |
| 4 | **안전한 에디터 제어** | 모든 커맨드에 명확한 피드백. 실수로 프로젝트를 깨뜨리지 않음. (`exec`는 명시적 예외 — 아래 참조) |
| 5 | **자동화 친화적** | JSON 출력, CI/CD 호환, AI 코딩 도구와 통합 가능. |

> **exec 예외:** `exec` 커맨드는 임의 GDScript를 실행하므로 원칙 4의 명시적 예외. opt-in 활성화 + 토큰 인증으로 위험을 관리한다.

---

## 2. 아키텍처

### 컴포넌트 다이어그램

```
┌──────────────────┐         HTTP (JSON)          ┌─────────────────────────┐
│                  │  POST /api/{command}           │                         │
│   godot-cli      │  Authorization: Bearer {token} │   Godot Editor 4.3+     │
│   (Go 바이너리)   │ ─────────────────────────────▶│   (EditorPlugin)        │
│                  │ ◀─────────────────────────────│                         │
└──────────────────┘         JSON 응답              └─────────────────────────┘
        │                                                       │
        │ 읽기                                                   │ 쓰기 (10초 주기)
        ▼                                                       ▼
┌──────────────────┐                               ┌─────────────────────────┐
│ ~/.godot-cli/    │                               │  @tool GDScript         │
│ instances/       │◀── heartbeat (10초) ─────────│  ├── HTTP Server        │
│   {pid}.json     │    (token 포함)               │  ├── Token Auth         │
└──────────────────┘                               │  ├── Tool Registry      │
        │                                          │  └── Command Handlers   │
        ▼                                          └─────────────────────────┘
  PID 존재 확인                                                  │
  (좀비 파일 삭제)                                               ▼
                                                      EditorSettings
                                                      godot_cli/enable_exec
```

### 동작 흐름

```
1. 사용자: godot-cli run --scene main.tscn

2. CLI: ~/.godot-cli/instances/ 파일 목록 읽기
3. CLI: PID 존재 확인으로 좀비 필터링
4. CLI: project_path 매칭 인스턴스 선택 → 토큰 추출

5. CLI:  POST http://localhost:{port}/api/run
         Authorization: Bearer {token}
         {"scene": "main.tscn"}

6. Plugin: 토큰 검증 → 불일치 시 401
7. Plugin: tool_registry에서 "run" 조회 → 없으면 404
8. Plugin: EditorInterface.play_custom_scene("res://main.tscn")
9. Plugin: {"success": true, "data": {"scene": "main.tscn", "running": true}}

10. CLI: 결과 출력 (종료 코드 0)
```

---

## 3. 기술 스택

### 결정 요약

| 항목 | 결정 | 탈락 대안 |
|------|------|-----------|
| **CLI 언어** | Go | Rust (오버스펙), Python (런타임 의존성) |
| **Godot 플러그인** | GDScript | C# (.mono 빌드 필수로 사용자 범위 축소) |
| **통신 프로토콜** | TCP + 간이 HTTP | WebSocket (단발 요청에 과잉), gRPC (GDScript 미지원) |
| **인스턴스 발견** | 파일 레지스트리 | 포트 스캔 (느림), 환경 변수 (다중 인스턴스 불가) |
| **Godot 최소 버전** | **4.3+** | — |

### Go (CLI)

- **cobra**: 커맨드 정의 및 플래그 파싱
- **표준 라이브러리**: HTTP 클라이언트, 파일 I/O, JSON
- **크로스컴파일**: Linux / macOS / Windows 단일 바이너리

### GDScript (Godot 플러그인)

- **EditorPlugin**: 에디터 수명 동안 실행, 에디터 API 전체 접근
- **TCPServer**: HTTP 서버 구현 (`_process()` 폴링 방식)
- **Crypto**: 보안 토큰 생성 (`Crypto.new().generate_random_bytes(16)`)

---

## 4. 기능 명세

### MVP v0.1 — 7개 커맨드

| 커맨드 | 설명 | Godot API |
|--------|------|-----------|
| `status` | 연결 상태, 프로젝트명, 에디터 버전 | `Engine.get_version_info()`, `ProjectSettings`, `EditorInterface` |
| `run` | 현재 씬 또는 지정 씬 실행 | `EditorInterface.play_current_scene()` / `play_custom_scene(path)` |
| `stop` | 실행 중인 씬 중지 | `EditorInterface.stop_playing_scene()` |
| `scene` | 현재 열린 씬 정보, 씬 전환 | `EditorInterface.get_edited_scene_root()`, `open_scene_from_path()` |
| `exec` | 임의 GDScript 실행 (opt-in 필요) | `GDScript.new().source_code` + 실행 노드 attach |
| `list` | 사용 가능한 커맨드 목록 | CLI 자체 메타데이터 (서버 불필요) |
| `connect` | 에디터 인스턴스 연결 확인 | 인스턴스 파일 읽기 + ping |

### exec 커맨드 상세

```bash
# 기본 사용 (자동 래핑 적용)
godot-cli exec "return Engine.get_version_info()"

# 멀티라인
godot-cli exec "var n = Node.new()
return n.get_class()"
```

**구현 방식 — `GDScript.new().source_code`**

CLI(또는 플러그인)가 사용자 코드를 아래 형태로 자동 래핑:

```gdscript
extends Node
func _cli_execute():
    {사용자 코드 (들여쓰기 적용)}
```

실행 흐름:
```gdscript
var script = GDScript.new()
script.source_code = _wrap_code(user_code)

var err = script.reload()
if err != OK:
    return _error("INVALID_PARAMS", "컴파일 에러", {"error_code": err})

var runner = Node.new()
runner.set_script(script)
EditorInterface.get_base_control().add_child(runner)

var result = runner._cli_execute()
var serialized = _serialize_variant(result)  # 직렬화 먼저
runner.queue_free()                           # 이후 정리

return _success(serialized)
```

> **경고:** exec 타임아웃은 HTTP 응답 수준에서만 동작합니다. GDScript는 단일 스레드이므로 무한 루프 코드 실행 시 에디터가 프리즈될 수 있습니다. 사용자 책임으로 실행하세요.

**Variant → JSON 직렬화 전략:**

| GDScript 타입 | JSON 표현 |
|--------------|-----------|
| `int`, `float`, `bool`, `String` | 직접 매핑 |
| `Dictionary`, `Array` | 자연 직렬화 |
| `Vector2` | `{"type": "Vector2", "x": 1.0, "y": 2.0}` |
| `Vector3` | `{"type": "Vector3", "x": 1.0, "y": 2.0, "z": 3.0}` |
| `Object` / `Node` | `{"type": "Object", "class": "Node", "id": instance_id}` |
| `null` | `null` |

### v0.2 — 개발 생산성

| 커맨드 | 설명 |
|--------|------|
| `log` | 에디터 출력 로그 읽기/필터링 (로그 파일 테일링) |
| `test` | GdUnit4/GUT 테스트 실행 |
| `export` | 프로젝트 내보내기 (`godot --headless --export-release` 래핑) |
| `screenshot` | 에디터/게임 뷰 PNG 캡처 |
| `node` | 씬 트리 탐색/조작 |
| `resource` | 리소스 검색/정보 조회 |
| `script` | GDScript 파일 직접 실행 |
| `init` | 플러그인 자동 설치 부트스트래핑 |

### v0.3 — 고급 자동화

| 커맨드 | 설명 |
|--------|------|
| `profiler` | 성능 프로파일러 데이터 조회 |
| `plugin` | 에디터 플러그인 활성화/비활성화 |
| `setting` | 프로젝트 설정 읽기/쓰기 |
| `update` | godot-cli 자체 업데이트 |
| `watch` | 파일 변경 감시 & 자동 리로드 |

---

## 5. 통신 프로토콜

### HTTP 서브셋 명세

- **메서드:** `POST`만 지원
- **경로:** `/api/{command}`
- **Content-Type:** `application/json`
- **Connection:** `close` 강제 (keep-alive 미지원)
- **Content-Length:** 필수 (chunked encoding 미지원)
- **인증:** `Authorization: Bearer {token}` 필수
- **최대 요청 크기:** 구현 시 결정 (OOM 방지 상한 설정 필요)

### 요청 형식

```http
POST /api/run HTTP/1.1
Host: localhost:8090
Authorization: Bearer a1b2c3d4e5f6...
Content-Type: application/json
Content-Length: 26
Connection: close

{"scene": "main.tscn"}
```

### 응답 형식

**성공:**
```json
{
  "success": true,
  "data": {
    "scene": "main.tscn",
    "running": true
  }
}
```

**에러:**
```json
{
  "success": false,
  "error": {
    "code": "COMMAND_NOT_FOUND",
    "message": "알 수 없는 커맨드: foo",
    "details": {}
  }
}
```

### 향후 프로토콜 교체 대비

Go CLI는 `Transport` 인터페이스로 추상화되어 있어, v0.3에서 WebSocket Transport를 추가해도 상위 레이어 변경이 없다:

```go
type Transport interface {
    Send(ctx context.Context, command string, params map[string]any) (*Response, error)
    Ping(ctx context.Context) error
    Close() error
}
```

---

## 6. 인스턴스 관리

### 인스턴스 파일 형식

**경로:** `~/.godot-cli/instances/{pid}.json`

```json
{
  "pid": 12345,
  "port": 8090,
  "project_path": "/home/user/my-godot-game",
  "godot_version": "4.4.1",
  "token": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
  "started_at": "2026-05-12T10:30:00Z",
  "last_heartbeat": "2026-05-12T10:35:00Z"
}
```

### 인스턴스 발견 알고리즘 (2단계)

```
1단계 — 파일 레지스트리 필터링
  1. ~/.godot-cli/instances/ 파일 목록 읽기
  2. 각 파일의 PID가 OS에 존재하는지 확인
     - Unix:   process.Signal(syscall.Signal(0))  → 에러 없으면 존재
     - Windows: windows.OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, ...)
  3. 좀비 파일(PID 미존재) 자동 삭제

2단계 — TCP 활성 확인
  4. 남은 인스턴스에 GET /api/ping 전송 (Bearer 토큰 포함)
  5. 응답 없거나 401이면 stale로 제외

인스턴스 선택
  6. project_path가 현재 디렉토리(또는 상위 Godot 프로젝트 루트)와 매칭되는 것 선택
  7. 다중 매칭 시: last_heartbeat가 가장 최근인 인스턴스 선택
  8. 매칭 없음 시: 종료 코드 2 + 안내 메시지
```

### Heartbeat

| 항목 | 값 |
|------|-----|
| Godot 플러그인 쓰기 주기 | **10초** |
| CLI stale 판정 기준 | **30초** 이상 |
| 파일 쓰기 방식 | 임시 파일 쓰기 후 rename (원자적, Windows 포함) |
| 비정상 종료 처리 | PID 기반 좀비 탐지 (heartbeat보다 빠른 1차 방어선) |

---

## 7. 보안 모델

### 토큰 인증

- Godot 플러그인이 시작 시 **32자 hex 토큰** 생성
  ```gdscript
  var crypto = Crypto.new()
  var token = crypto.generate_random_bytes(16).hex_encode()
  ```
- 토큰은 인스턴스 파일(`~/.godot-cli/instances/{pid}.json`)에 저장
- 모든 HTTP 요청에 `Authorization: Bearer {token}` 헤더 필수
- 토큰 불일치 또는 누락 시 **HTTP 401** 반환

### exec 보안

- **기본 비활성화:** `godot_cli/enable_exec = false` (EditorSettings)
- 활성화 방법: **Project Settings → Plugins → godot_cli → enable_exec = true**
- 비활성화 상태에서 exec 호출 시 **HTTP 403 (`EXEC_DISABLED`)** 반환
- 로컬호스트(`127.0.0.1`)에만 바인딩 — 원격 접근 차단

### 공격 표면

```
외부 네트워크  → localhost 바인딩으로 차단
같은 머신 프로세스 → 토큰 인증으로 차단 (인스턴스 파일은 사용자 소유)
exec 악용 → opt-in + 토큰 필수 + 향후 감사 로그
```

---

## 8. 디렉토리 구조

### Go CLI

```
godot-cli/
├── cmd/
│   ├── root.go          # 루트 커맨드, 글로벌 플래그 (--port, --project, --json)
│   ├── status.go
│   ├── run.go
│   ├── stop.go
│   ├── scene.go
│   ├── exec.go
│   ├── list.go
│   └── connect.go
├── internal/
│   ├── client/
│   │   ├── transport.go       # Transport 인터페이스
│   │   ├── http_transport.go  # HTTP 구현체 (Bearer 토큰 자동 첨부)
│   │   ├── client.go          # Transport를 사용하는 Client
│   │   └── discovery.go       # 인스턴스 발견 로직
│   ├── instance/
│   │   ├── registry.go        # 인스턴스 파일 읽기 + 토큰 추출
│   │   ├── heartbeat.go       # 30초 stale 판정
│   │   └── zombie.go          # PID 기반 좀비 탐지 (플랫폼별 분기)
│   ├── output/
│   │   ├── formatter.go       # text/json 출력 전환
│   │   └── table.go           # 테이블 출력
│   └── errors/
│       └── errors.go          # 에러 코드 + 종료 코드 정의
├── main.go
├── go.mod
├── Makefile
└── README.md
```

### Godot 플러그인

```
addons/godot_cli/
├── plugin.cfg                 # godot_version = "4.3"
├── plugin.gd                  # EditorPlugin 메인
├── server/
│   ├── http_server.gd         # TCPServer 기반 HTTP 서버 (_process() 폴링)
│   ├── request_parser.gd      # HTTP 요청 파싱
│   ├── response_builder.gd    # HTTP 응답 생성 (에러 스키마 준수)
│   └── auth.gd                # Bearer 토큰 검증 미들웨어
├── tools/
│   ├── tool_registry.gd       # 도구 등록/발견
│   ├── base_tool.gd           # 베이스 클래스 + _success()/_error() 헬퍼
│   ├── status_tool.gd
│   ├── run_tool.gd
│   ├── scene_tool.gd
│   └── exec_tool.gd           # GDScript.new() + Expression 보조
├── instance/
│   ├── instance_manager.gd    # 인스턴스 파일 생성/갱신/삭제 + 토큰 생성
│   └── heartbeat.gd           # 10초 타이머
└── util/
    ├── json_helper.gd
    └── variant_serializer.gd  # Variant → JSON 직렬화
```

---

## 9. 에러 처리

### 에러 코드

| 에러 코드 | HTTP 상태 | 설명 |
|-----------|-----------|------|
| `COMMAND_NOT_FOUND` | 404 | 요청한 커맨드 없음 |
| `INVALID_PARAMS` | 400 | 파라미터 누락/형식 오류, GDScript 컴파일 에러 포함 |
| `EXEC_DISABLED` | 403 | exec 커맨드 비활성화 상태 |
| `UNAUTHORIZED` | 401 | 토큰 누락 또는 불일치 |
| `COMMAND_TIMEOUT` | 408 | 커맨드 실행 타임아웃 |
| `EDITOR_ERROR` | 500 | 에디터 API 호출 실패 |
| `INTERNAL_ERROR` | 500 | 서버 내부 오류 |

### Go CLI 종료 코드

| 코드 | 의미 |
|------|------|
| `0` | 성공 |
| `1` | 커맨드 에러 (서버가 에러 응답 반환) |
| `2` | 연결 실패 (인스턴스 없음, 네트워크 에러) |
| `3` | 인증 실패 (토큰 불일치) |

### base_tool.gd 헬퍼

```gdscript
class_name GodotCliTool
extends RefCounted

func get_name() -> String: return ""
func get_description() -> String: return ""
func get_parameters() -> Array[Dictionary]: return []
func execute(params: Dictionary) -> Dictionary:
    return _error("INTERNAL_ERROR", "Not implemented")

func _success(data: Variant = null) -> Dictionary:
    return {"success": true, "data": data}

func _error(code: String, message: String, details: Dictionary = {}) -> Dictionary:
    return {"success": false, "error": {"code": code, "message": message, "details": details}}
```

---

## 10. 구현 단계

### Step 1: Godot HTTP 서버 기반 + 토큰 인증

**목표:** 토큰 인증된 HTTP 요청을 받고 통일된 JSON 응답을 반환하는 최소 서버

**할 일:**
- [ ] `plugin.cfg` + `plugin.gd` EditorPlugin 스켈레톤 (`godot_version = "4.3"`)
- [ ] `TCPServer` 기반 HTTP 서버 (`_process()` 폴링 방식)
- [ ] HTTP 요청 파서 (메서드, 경로, 헤더, JSON 바디, Authorization 헤더)
- [ ] HTTP 응답 빌더 (에러 스키마 준수)
- [ ] `auth.gd` Bearer 토큰 검증 (불일치 시 401)
- [ ] `/api/ping` 엔드포인트

**완료 기준:**
```bash
# 토큰 있을 때
curl -H "Authorization: Bearer {token}" -X POST http://localhost:{port}/api/ping
# → {"success": true, "data": {"pong": true}}

# 토큰 없을 때
curl -X POST http://localhost:{port}/api/ping
# → {"success": false, "error": {"code": "UNAUTHORIZED", ...}}  (HTTP 401)

# 없는 커맨드
curl -H "Authorization: Bearer {token}" -X POST http://localhost:{port}/api/foo
# → {"success": false, "error": {"code": "COMMAND_NOT_FOUND", ...}}  (HTTP 404)
```

---

### Step 2: 인스턴스 관리

**목표:** 보안 토큰 포함 다중 에디터 인스턴스 발견/연결

**할 일:**
- [ ] 인스턴스 파일에 `token` 필드 추가 (`Crypto.new().generate_random_bytes(16).hex_encode()`)
- [ ] `~/.godot-cli/instances/{pid}.json` 생성/갱신/삭제
- [ ] **10초** 하트비트 타이머 (Windows: 임시 파일 + rename 원자적 쓰기)
- [ ] 동적 포트 할당 (8090부터 순차 시도)
- [ ] `_exit_tree()`에서 인스턴스 파일 정리
- [ ] EditorSettings에 `godot_cli/enable_exec` 설정 등록 (기본값: false)

**완료 기준:**
- 두 Godot 에디터를 열면 각각 다른 포트 + 다른 토큰의 인스턴스 파일 생성
- 에디터 종료 시 파일 삭제
- 하트비트가 10초 간격으로 `last_heartbeat` 갱신

---

### Step 3: Go CLI 스켈레톤

**목표:** cobra 기반 CLI + Transport 인터페이스 + 인스턴스 발견

**할 일:**
- [ ] Go 모듈 초기화, cobra 의존성
- [ ] `internal/client/transport.go` — Transport 인터페이스
- [ ] `internal/client/http_transport.go` — HTTP 구현체 (Bearer 자동 첨부)
- [ ] `internal/client/client.go` — Client
- [ ] `internal/instance/registry.go` — 파일 읽기 + 토큰 추출
- [ ] `internal/instance/zombie.go` — PID 기반 좀비 탐지 (Unix/Windows 분기)
- [ ] `internal/instance/heartbeat.go` — 30초 stale 판정
- [ ] `internal/errors/errors.go` — 에러 코드 + 종료 코드
- [ ] `cmd/root.go` — 글로벌 플래그
- [ ] `cmd/status.go`, `cmd/list.go`, `cmd/connect.go`

**완료 기준:**
```bash
godot-cli status          # 에디터 정보 출력
godot-cli status --json   # JSON 출력
# 종료 코드: 인증 실패 3, 연결 실패 2
```

---

### Step 4: MVP 커맨드 구현

**목표:** 7개 MVP 커맨드 완성

**할 일:**
- [ ] `GodotCliTool` 베이스 클래스 + `ToolRegistry`
- [ ] `variant_serializer.gd` — Variant → JSON
- [ ] `status_tool.gd`, `run_tool.gd`, `scene_tool.gd`
- [ ] `exec_tool.gd` — 자동 래핑 + GDScript.new() + enable_exec 체크
- [ ] Go CLI: `run`, `stop`, `scene`, `exec` 커맨드

**완료 기준:**
```bash
godot-cli run                        # 현재 씬 실행
godot-cli run --scene res://main.tscn
godot-cli stop
godot-cli scene
godot-cli exec "return Engine.get_version_info()"   # opt-in 활성화 시
godot-cli exec "var x = 42\nreturn x"               # 멀티라인
```
- exec: `enable_exec = false`일 때 `EXEC_DISABLED` (HTTP 403)
- exec: 컴파일 에러 시 `INVALID_PARAMS` + 에러 메시지
- 모든 커맨드: `--json` 플래그로 JSON 출력

---

### Step 5: 품질 & 배포

**목표:** 안정성 확보, 크로스플랫폼 빌드, 문서화

**할 일:**
- [ ] Go 유닛 테스트 (transport, client, instance, zombie, 각 커맨드)
- [ ] 에러 코드 7종 × 종료 코드 4종 통합 테스트
- [ ] Makefile / GoReleaser (Linux, macOS, Windows)
- [ ] README.md (설치, 플러그인 설치, 사용법, exec 활성화, 보안 설정)
- [ ] Godot 플러그인 AssetLib 형태 정리

**완료 기준:**
- `goreleaser` 3개 플랫폼 빌드 성공
- README에 설치 → 첫 커맨드까지 완전한 가이드
- exec 경고(단일 스레드 한계) README에 포함

---

## 11. 성공 기준

### 기능 기준

- [ ] `godot-cli status` — 에디터 정보 출력
- [ ] `godot-cli run` — 현재 씬 실행
- [ ] `godot-cli stop` — 씬 중지
- [ ] `godot-cli scene` — 씬 정보 출력
- [ ] `godot-cli exec "return Engine.get_version_info()"` — GDScript 실행 (enable_exec 활성화 시)
- [ ] `godot-cli list` — 커맨드 목록
- [ ] `godot-cli connect` — 인스턴스 연결 확인

### 인프라 기준

- [ ] 다중 에디터 인스턴스에서 올바른 인스턴스 선택
- [ ] `--json` 플래그로 모든 커맨드의 JSON 출력
- [ ] Windows / macOS / Linux 동작
- [ ] 에디터 정상 종료 시 인스턴스 파일 삭제 (좀비 없음)
- [ ] PID 기반 좀비 파일 자동 정리

### 보안 기준

- [ ] 모든 요청에 토큰 인증 적용
- [ ] 토큰 없는 요청 → HTTP 401
- [ ] exec 기본 비활성화, EditorSettings에서 opt-in
- [ ] exec 비활성화 상태 호출 → HTTP 403

### 품질 기준

- [ ] 에러 응답 스키마 7종 에러 코드 준수
- [ ] Go CLI 종료 코드 (0/1/2/3) 정확한 매핑
- [ ] Godot 4.3+ 지원 확인
- [ ] 하트비트 10초, stale 판정 30초

---

## 12. 위험 요소

### 높은 위험

| 위험 | 완화 전략 | 상태 |
|------|-----------|------|
| **TCPServer 메인 스레드 블로킹** | `_process()`에서 `is_connection_available()` 폴링, 처리 분산 | 미검증 |
| **HTTP 파싱 정확성** | `Connection: close` 강제, Content-Length 필수, 최소 서브셋 지원 | 전략 확정 |
| **EditorInterface API 제한** | 각 API 호출 전 상태 검증, 철저한 에러 핸들링 | 미검증 |

### 중간 위험

| 위험 | 완화 전략 | 상태 |
|------|-----------|------|
| **exec 보안** | 토큰 인증 + 기본 비활성화 + opt-in + 향후 감사 로그 | 설계 완료 |
| **exec 메모리 누수** | `_serialize_variant()` 완료 후 `queue_free()`, 타임아웃 | 전략 확정 |
| **Windows 파일 원자성** | 임시 파일 + rename (같은 볼륨, 원자적) | 전략 확정 |

### 미검증 항목 (Step 1~2에서 프로토타입 검증 필요)

| 항목 | 질문 |
|------|------|
| **EditorPlugin 라이프사이클** | 씬 실행(F5) 중에도 HTTP 서버를 유지하는가? |
| **TCPServer.listen(0)** | OS에서 동적 포트를 할당받을 수 있는가? |
| **exec 노드 씬 트리 접근** | `get_base_control().add_child()` 노드에서 EditorInterface 접근되는가? |

---

## 13. 로드맵

```
v0.1 (MVP)     status, run, stop, scene, exec, list, connect
               보안 토큰, 에러 스키마, 3개 플랫폼 빌드

v0.2           log, test, export, screenshot, node, resource, script
               godot-cli init (플러그인 자동 설치)
               exec 감사 로그

v0.3           profiler, plugin, setting, update, watch
               WebSocket Transport (이벤트 스트리밍)
               도구 자동 발견 (디렉토리 스캔)
```

---

## 14. 의식적 절충 기록

### 절충 1: HTTP vs GDScript 관용구

- **상황:** HTTP 통신은 GDScript 생태계의 일반 패턴이 아님 (WebSocket이 Godot 1등 시민)
- **선택 근거:** curl/브라우저 직접 디버깅 용이, godot-cli에서 검증된 패턴, REST 스타일 커맨드 매핑의 자연스러움
- **완화:** Transport 인터페이스로 향후 WebSocket 전환 비용 최소화

### 절충 2: exec 보안 vs 자동화 편의성

- **상황:** 임의 코드 실행은 원칙 4 "안전한 에디터 제어"와 상충
- **선택 근거:** AI 코딩 도구, CI/CD 자동화에서 exec는 핵심 가치
- **완화:** opt-in 활성화 + 토큰 인증 필수 + 향후 감사 로그

### 절충 3: 하트비트 간격 10초

- **상황:** 비정상 종료 시 최대 30초간 stale 인스턴스가 목록에 남음
- **선택 근거:** 1초는 파일 I/O 부담, PID 기반 좀비 탐지가 빠른 1차 방어선
- **완화:** PID 체크 + TCP ping 2단계 발견으로 false positive 최소화

---

*이 설계서는 Planner → Architect → Critic 합의 루프(2회 이터레이션)를 통해 작성되었습니다.*
