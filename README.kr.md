# godot-cli

[English](README.md) | [한국어](README.kr.md)

> 커맨드라인에서 Godot 에디터를 제어합니다.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Go 단일 바이너리와 가벼운 GDScript 애드온의 조합. 에디터 UI를 거치지 않고 터미널에서 씬 실행, 씬 트리 탐색, 프로젝트 설정 편집, 스크린샷 캡처, 빌드 내보내기, 유닛 테스트 실행 등을 수행할 수 있습니다.

## 요구사항

- Godot **4.3+** (표준 빌드, `.mono` 불필요)
- Go 1.22+ (소스에서 빌드할 때만)

## 설치

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/Sqoff/godot-cli/main/install.ps1 | iex
```

### Linux / macOS

```bash
curl -fsSL https://github.com/Sqoff/godot-cli/releases/latest/download/godot-cli-linux-amd64 -o godot-cli
chmod +x godot-cli && sudo mv godot-cli /usr/local/bin/
```

### 소스에서 빌드

```bash
git clone https://github.com/Sqoff/godot-cli.git
cd godot-cli && make build
```

지원 플랫폼: Linux (amd64, arm64), macOS (Intel, Apple Silicon), Windows (amd64).

### 업데이트

```bash
godot-cli update
```

## Godot 설정

Godot 프로젝트 루트에서 애드온을 설치합니다.

```bash
godot-cli init --enable      # 애드온 복사 + project.godot에 자동 활성화
# 또는 복사만 하고 Project Settings에서 직접 활성화:
godot-cli plugin install
```

Godot에서 프로젝트를 열면 Output 패널에 다음이 출력됩니다.

```
[godot-cli] Server running on port 8090
[godot-cli] Token: <bearer token>
```

설정은 이게 전부입니다. 애드온이 프로젝트별 인스턴스 파일을 `~/.godot-cli/instances/{pid}.json` (포트, 토큰, 프로젝트 경로, 하트비트)에 기록해 CLI가 올바른 에디터를 찾아갑니다.

CLI를 업데이트한 뒤 애드온을 갱신하려면:

```bash
godot-cli plugin install --force
```

## 빠른 시작

```bash
godot-cli status
godot-cli run
godot-cli stop
godot-cli node tree --depth 2
godot-cli log --lines 30 --type error
```

## 동작 원리

```
터미널                                Godot 에디터
──────                                ────────────
$ godot-cli run --scene main.tscn
    │
    ├─ ~/.godot-cli/instances/*.json 읽기
    │  → 현재 작업 디렉토리로 매칭
    │  → 포트와 bearer 토큰 추출
    │
    ├─ POST http://127.0.0.1:8090/api/run
    │  Authorization: Bearer <token>
    │  { "scene": "main.tscn" }
    │                                      │
    │                              TCPServer 수락 (auth.gd가 토큰 검증)
    │                                      │
    │                              tool_registry가 run_tool.gd로 디스패치
    │                                      │
    │                              EditorInterface.play_custom_scene(...)
    │                                      │
    ├─ JSON 응답 ←──────────────────────────┘
    │  { "success": true, "data": { "started": true } }
    │
    └─ 출력: Game started  scene=main.tscn
```

- 애드온은 `_process()`에서 폴링하는 `TCPServer`를 소유한 `@tool` GDScript `EditorPlugin`입니다.
- 도구는 `addons/godot_cli/tools/`에 있고 파일명으로 자동 발견됩니다 — `<name>_tool.gd`는 다음 에디터 리로드 시 `<name>` 커맨드로 등록됩니다.
- 모든 요청은 프로젝트별 bearer 토큰으로 인증됩니다.
- 10초 하트비트가 인스턴스 파일을 갱신하고, CLI는 매 조회 전에 PID가 사라진 좀비 파일을 청소합니다.

## 빌트인 커맨드

| 커맨드 | 설명 |
|--------|------|
| `status` | 프로젝트 경로, Godot 버전, 현재 씬 표시 |
| `connect` | 에디터 연결 확인 |
| `list` | 에디터에 등록된 커맨드 목록 (또는 `--local`로 CLI 측 목록) |
| `run [--scene <p>]` | 메인 씬 또는 지정 씬 실행 |
| `stop` | 실행 중인 씬 중지 |
| `pause [--on \| --off]` | 실행 중인 씬 일시정지/재개/토글 |
| `scene` | 현재 씬 정보 표시 |
| `refresh [--sources]` | 에셋 파일시스템 재스캔 |
| `menu <action>` | 화이트리스트된 에디터 액션 실행 |
| `log` | 최근 에디터 로그 읽기 (필터 지원) |
| `exec <node> <method> [args…]` | 씬 노드의 메서드 호출 (opt-in) |
| `script <file>` | 에디터 컨텍스트에서 GDScript 파일 실행 (opt-in) |
| `node tree \| get \| set` | 편집 중인 씬 트리 탐색/수정 |
| `resource find \| info` | 프로젝트 리소스 검색/조회 |
| `reserialize` | `.tscn`/`.tres`를 Godot 직렬화기로 다시 저장 |
| `setting get \| set \| list` | `ProjectSettings` 읽기/쓰기 |
| `profiler status \| enable \| disable \| clear \| hierarchy` | 성능 모니터 |
| `screenshot --output <p>` | 에디터 뷰포트를 PNG로 캡처 |
| `watch` | 파일 변경 감시 + 재스캔 트리거 |
| `export --preset --output` | `godot --headless --export-release/-debug` 래핑 |
| `test` | GUT (Godot Unit Test) 실행 |
| `init` | 현재 프로젝트에 애드온 부트스트랩 |
| `plugin install` | 애드온만 설치/갱신 |
| `update` | CLI 바이너리 자체 업데이트 |

모든 커맨드는 구조화된 출력을 위한 `--json` 플래그를 지원합니다.

### 에디터 제어

```bash
godot-cli run                     # 메인 씬 실행
godot-cli run --scene res://levels/boss.tscn
godot-cli stop

godot-cli pause                   # 토글
godot-cli pause --on              # 강제 일시정지
godot-cli pause --off             # 강제 재개

godot-cli refresh                 # 전체 파일시스템 스캔
godot-cli refresh --sources       # 소스 파일만

godot-cli menu scene/save
godot-cli menu scene/save_all
godot-cli menu scene/reload
godot-cli menu filesystem/scan
```

`menu`는 의도적으로 화이트리스트 방식입니다 — 임의의 에디터 메뉴 경로는 호출되지 않습니다.

### 에디터 로그

```bash
godot-cli log                                # 최근 50줄 (기본값)
godot-cli log --lines 200 --type error
godot-cli log --type warn
godot-cli log --filter "ResourceLoader"
godot-cli log --clear                        # 읽은 뒤 로그 파일 비우기
```

애드온이 `OS.get_user_data_dir()/logs/` 아래의 로그 파일을 찾아 `ERROR` / `WARNING` 마커로 필터링합니다.

### exec — 노드 메서드 호출

```bash
godot-cli exec . get_name                    # 씬 루트의 get_name()
godot-cli exec Player get_hp
godot-cli exec UI/HUD update_score 42        # JSON 파싱 가능한 인자는 타입 전달
godot-cli exec . get_name --json
```

노드 경로는 현재 편집 중인 씬의 루트 기준 상대 경로입니다. JSON 파싱 가능한 인자(`42`, `true`, `"foo"`, `{"k":1}`)는 타입으로 전달되고 나머지는 문자열로 처리됩니다.

**보안:** `exec`는 기본 비활성화. **에디터 → 에디터 설정 → `godot_cli/enable_exec` = true**로 활성화해야 합니다.

### script — GDScript 파일 실행

단일 메서드 호출로 부족할 때 멀티라인 로직을 실행:

```gdscript
# /tmp/audit.gd
extends Node
func _cli_execute():
    var root := EditorInterface.get_edited_scene_root()
    return {"name": root.name, "child_count": root.get_child_count()}
```

```bash
godot-cli script /tmp/audit.gd
godot-cli script /tmp/audit.gd --method _cli_execute --json
```

`extends`가 없는 짧은 스니펫은 자동 래핑됩니다 — 애드온이 `extends Node` + `func _cli_execute():`를 앞에 붙이고 `GDScript.new().source_code`로 컴파일, 임시 `Node`에 attach 후 진입 메서드를 호출하고 정리합니다. `exec`와 동일한 `enable_exec` 게이트가 적용됩니다.

### 씬 트리

```bash
godot-cli node tree                          # 씬 루트부터 깊이 3
godot-cli node tree Player --depth 1 --props
godot-cli node get Player/Sprite scale
godot-cli node set Player/Sprite scale '{"x":2,"y":2}'
```

`node set`은 `enable_exec`가 필요합니다.

### 리소스

```bash
godot-cli resource find boss                          # 경로 부분 일치 검색
godot-cli resource info res://scenes/main.tscn

godot-cli reserialize --path res://scenes/main.tscn   # 단일 파일
godot-cli reserialize --all --dry-run                 # 대상 미리보기
godot-cli reserialize --all                           # 실제 저장
```

`reserialize`는 각 `.tscn`/`.tres`를 `ResourceLoader`로 읽어 `ResourceSaver`로 다시 씁니다. 씬/리소스 파일을 텍스트로 일괄 편집한 뒤 정리할 때 유용합니다.

### 프로젝트 설정

```bash
godot-cli setting get application/config/name
godot-cli setting list                                # 사용자 정의 항목
godot-cli setting list --prefix myaddon/
godot-cli setting list --all                          # 빌트인 포함
godot-cli setting set myaddon/threshold 0.5
```

`setting set`은 `ProjectSettings.save()`를 호출해 디스크의 `project.godot`을 수정하므로 `enable_exec`가 필요합니다.

### 프로파일러

```bash
godot-cli profiler status
godot-cli profiler enable                             # on-demand 샘플 수집 시작
godot-cli profiler hierarchy                          # 최신 캡처
godot-cli profiler hierarchy --frame 3 --top 20
godot-cli profiler clear
godot-cli profiler disable
```

`Performance` 싱글톤을 캡처합니다 — FPS, 프레임/물리 시간, 메모리, 오브젝트/노드/orphan 수, draw call 등.

### 스크린샷

```bash
godot-cli screenshot --output /tmp/editor.png
godot-cli screenshot --output /tmp/game.png --target game
```

애드온이 에디터 base control의 뷰포트를 캡처해 PNG로 인코딩, base64로 CLI에 스트리밍하면 CLI가 파일로 저장합니다.

### Watch

```bash
godot-cli watch --once                                # 한 번 스냅샷 후 종료
godot-cli watch --path res://addons/ --interval 2     # 폴링하며 변경 감지 시 재스캔
```

### Export

```bash
godot-cli export --preset "Windows Desktop" --output build/game.exe
godot-cli export --preset "Linux/X11"      --output build/game.x86_64 --debug
godot-cli export --preset Web              --output build/index.html --godot /opt/godot/godot
```

CLI 전용 — `godot --headless --export-release` (또는 `--export-debug`) 래핑. PATH의 `godot` / `godot4` 자동 감지하며 `--godot <path>`로 덮어쓸 수 있습니다.

### 테스트 (GUT)

```bash
godot-cli test                                # res://test/ 기본
godot-cli test --directory res://tests/unit/
godot-cli test --filter test_player_
godot-cli test --mode editor                  # spawn 대신 GUT 패널 안내
godot-cli test --godot /opt/godot/godot
```

[GUT 애드온](https://github.com/bitwes/Gut)이 `res://addons/gut/`에 설치되어 있어야 합니다. standalone 모드는 `godot --headless -s addons/gut/gut_cmdln.gd -gdir=<dir> -gexit`를 실행합니다.

### status, list, connect

```bash
godot-cli status
godot-cli connect
godot-cli list                                # 에디터에 등록된 도구 조회
godot-cli list --local                        # 에디터 없이 CLI 측 목록만
```

## 전역 옵션

| 플래그 | 설명 |
|--------|------|
| `--port <n>` | 특정 에디터 인스턴스를 포트로 지정 |
| `--project <path>` | 특정 프로젝트 디렉토리로 지정 |
| `--json` | 원시 JSON 출력 |

```bash
godot-cli --port 8091 status
godot-cli --project ~/games/mygame run
godot-cli status --json
```

각 커맨드의 `--help`로 상세 사용법 확인:

```bash
godot-cli node tree --help
godot-cli profiler hierarchy --help
```

## 다중 에디터 인스턴스

여러 Godot 에디터가 동시에 열려 있으면 각각 다른 포트(8090, 8091, …)에 바인딩하고 자신의 인스턴스 파일을 작성합니다. CLI는 기본으로 현재 작업 디렉토리로 매칭하고, 명시적 `--port` 또는 `--project`가 우선합니다.

```bash
ls ~/.godot-cli/instances/
# 12345.json  12390.json

godot-cli --port 8091 status
godot-cli --project ~/games/mygame run
godot-cli status            # cwd 매칭; 인스턴스가 하나면 자동 선택
```

각 커맨드 직전에 PID가 살아있지 않은 인스턴스 파일은 제거되므로, 종료된 에디터가 잔존하지 않습니다.

## 커스텀 도구 작성

`addons/godot_cli/tools/` 디렉토리에 `<name>_tool.gd` 파일을 추가하기만 하면 됩니다. registry가 에디터 시작 시 이 디렉토리를 스캔해 `<name>` 커맨드로 등록합니다. 등록 호출, 매니페스트, 재빌드 모두 불필요합니다.

```gdscript
# addons/godot_cli/tools/spawn_tool.gd
@tool
extends RefCounted

const ResponseBuilder := preload("res://addons/godot_cli/server/response_builder.gd")

func handle(request: Dictionary, _plugin: EditorPlugin) -> Dictionary:
    var body = JSON.parse_string(request.get("body", ""))
    if not body is Dictionary:
        body = {}
    var prefab_path: String = body.get("prefab", "")
    if prefab_path.is_empty():
        return ResponseBuilder.error(400, "INVALID_PARAMS", "Missing 'prefab'")

    var packed: PackedScene = ResourceLoader.load(prefab_path)
    if packed == null:
        return ResponseBuilder.error(404, "RESOURCE_NOT_FOUND", "Prefab not found")

    var instance := packed.instantiate()
    instance.position = Vector2(float(body.get("x", 0)), float(body.get("y", 0)))
    var root := EditorInterface.get_edited_scene_root()
    root.add_child(instance)
    instance.owner = root

    return ResponseBuilder.success({"name": instance.name})
```

어디서든 호출 가능:

```bash
curl -X POST http://127.0.0.1:8090/api/spawn \
  -H "Authorization: Bearer <token>" \
  -d '{"prefab":"res://prefabs/goblin.tscn","x":10,"y":5}'
```

**규약:**

- 파일명 = 커맨드명. `spawn_tool.gd` → `spawn`. registry가 어떤 스크립트를 로드할지 식별하려면 `_tool.gd` 접미사가 필수입니다.
- 진입점: `func handle(request: Dictionary, _plugin: EditorPlugin) -> Dictionary`.
- `request` 딕셔너리에는 `path`, `headers`, `body`(원시 JSON 문자열)가 있습니다. `body`는 `JSON.parse_string`으로 파싱하세요.
- 응답은 `ResponseBuilder.success(data)` 또는 `ResponseBuilder.error(http_status, code, message, details?)`로 만듭니다.
- Godot 타입(`Vector2`, `Color`, `Object` 등)을 응답에 담을 때는 `util/variant_serializer.gd`를 거치게 해 깨끗한 JSON으로 변환하세요.
- 인증은 중앙에서 처리됩니다 — 도구는 bearer 토큰 검증 통과 후에만 실행됩니다.
- 함수명에 `_get`, `_set`, `_init` 같은 **Godot Object 가상 메서드 이름은 피하세요** — 시그니처 충돌로 parse 에러가 납니다. `_do_get`, `_handle_set` 같은 이름을 쓰는 것을 권장합니다.

타입이 명시된 Go 서브커맨드로 노출하려면 `cmd/run.go`나 `cmd/scene.go`를 모델로 삼아 `cmd/` 아래에 파일을 추가합니다 (`cmdutil.Resolve`로 인스턴스 해결, `client.NewHTTPTransport`로 전송, `--json`이 켜졌을 때 `output.PrintJSON`으로 출력).

## 종료 코드

| 코드 | 의미 |
|------|------|
| 0 | 성공 |
| 1 | 커맨드 에러 (stderr 확인) |
| 2 | 활성 Godot 에디터 인스턴스 없음 |
| 3 | 인증 실패 |

## 프로젝트 구조

```
godot-cli/
├── cmd/                          # cobra 서브커맨드 (커맨드당 .go 하나)
├── internal/
│   ├── client/                   # Transport 인터페이스 + HTTP 구현
│   ├── cmdutil/                  # 인스턴스 해결 헬퍼
│   ├── errors/                   # 에러 코드 + 종료 코드
│   ├── instance/                 # 하트비트 + 좀비 정리
│   └── output/                   # JSON / 텍스트 포매터
├── addons/godot_cli/
│   ├── plugin.cfg / plugin.gd    # EditorPlugin 진입점
│   ├── server/                   # TCPServer, 요청 파서, 인증, 응답 빌더
│   ├── instance/                 # 인스턴스 파일 + 하트비트
│   ├── tools/                    # *_tool.gd — 자동 발견되는 커맨드 핸들러
│   └── util/                     # variant 직렬화기
├── main.go
└── Makefile
```

## 빌드

```bash
make build                 # 현재 플랫폼 바이너리
make snapshot              # goreleaser로 모든 플랫폼 빌드
make release               # 태그 기반 릴리즈
make clean                 # 바이너리와 dist/ 정리
```

## 라이선스

MIT
