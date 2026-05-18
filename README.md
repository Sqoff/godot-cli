# godot-cli

[English](README.md) | [한국어](README.kr.md)

> Control the Godot editor from the command line.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A single Go binary plus a small GDScript addon. Run scenes, inspect the scene tree, edit project settings, capture screenshots, export builds, and run unit tests without touching the editor UI.

## Requirements

- Godot **4.3+** (standard build; `.mono` not required)
- Go 1.22+ (only when building from source)

## Install

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/Sqoff/godot-cli/main/install.ps1 | iex
```

### Linux / macOS

```bash
curl -fsSL https://github.com/Sqoff/godot-cli/releases/latest/download/godot-cli-linux-amd64 -o godot-cli
chmod +x godot-cli && sudo mv godot-cli /usr/local/bin/
```

### Build from source

```bash
git clone https://github.com/Sqoff/godot-cli.git
cd godot-cli && make build
```

Supported platforms: Linux (amd64, arm64), macOS (Intel, Apple Silicon), Windows (amd64).

### Update

```bash
godot-cli update
```

## Godot Setup

From your Godot project root, install the addon:

```bash
godot-cli init --enable      # copies the addon and enables it in project.godot
# or copy only, and enable manually in Project Settings:
godot-cli plugin install
```

Open the project in Godot. The addon prints to the Output panel:

```
[godot-cli] Server running on port 8090
[godot-cli] Token: <bearer token>
```

That's the whole setup. The addon writes a per-project instance file to `~/.godot-cli/instances/{pid}.json` (port, token, project path, heartbeat) so the CLI can find the right editor.

To refresh the addon after a CLI update:

```bash
godot-cli plugin install --force
```

## Quick Start

```bash
godot-cli status
godot-cli run
godot-cli stop
godot-cli node tree --depth 2
godot-cli log --lines 30 --type error
```

## How It Works

```
Terminal                              Godot Editor
────────                              ────────────
$ godot-cli run --scene main.tscn
    │
    ├─ reads ~/.godot-cli/instances/*.json
    │  → matches by current working directory
    │  → picks port and bearer token
    │
    ├─ POST http://127.0.0.1:8090/api/run
    │  Authorization: Bearer <token>
    │  { "scene": "main.tscn" }
    │                                      │
    │                              TCPServer accepts (auth.gd verifies token)
    │                                      │
    │                              tool_registry dispatches to run_tool.gd
    │                                      │
    │                              EditorInterface.play_custom_scene(...)
    │                                      │
    ├─ receives JSON ←─────────────────────┘
    │  { "success": true, "data": { "started": true } }
    │
    └─ prints: Game started  scene=main.tscn
```

- The addon is a `@tool` GDScript `EditorPlugin` that owns a `TCPServer` polled from `_process()`.
- Tools live in `addons/godot_cli/tools/` and are discovered by filename — any `<name>_tool.gd` becomes the `<name>` command on next editor reload.
- Every request is authenticated with the per-project bearer token.
- A 10-second heartbeat keeps the instance file fresh; the CLI deletes files whose PIDs no longer exist before each lookup.

## Built-in Commands

| Command | Description |
|---------|-------------|
| `status` | Show project path, Godot version, current scene |
| `connect` | Verify the editor connection |
| `list` | List commands registered in the editor (or CLI-side with `--local`) |
| `run [--scene <p>]` | Play the main scene or a specific scene |
| `stop` | Stop the running scene |
| `pause [--on \| --off]` | Pause, resume, or toggle the running scene |
| `scene` | Show or change the active scene |
| `refresh [--sources]` | Re-scan the asset filesystem |
| `menu <action>` | Invoke a whitelisted editor action |
| `log` | Read recent editor log lines with filters |
| `exec <node> <method> [args…]` | Call a method on a scene node (opt-in) |
| `script <file>` | Run a GDScript file in editor context (opt-in) |
| `node tree \| get \| set` | Inspect or mutate the edited scene tree |
| `resource find \| info` | Search and inspect project resources |
| `reserialize` | Re-save `.tscn`/`.tres` through Godot's serializer |
| `setting get \| set \| list` | Read or write `ProjectSettings` entries |
| `profiler status \| enable \| disable \| clear \| hierarchy` | Performance monitors |
| `screenshot --output <p>` | Capture the editor viewport as PNG |
| `watch` | Poll filesystem for changes and trigger rescans |
| `export --preset --output` | Wrap `godot --headless --export-release/-debug` |
| `test` | Run GUT (Godot Unit Test) suites |
| `init` | Bootstrap the addon into the current project |
| `plugin install` | Install or refresh the addon only |
| `update` | Self-update the CLI binary |

All commands support `--json` for structured output.

### Editor control

```bash
godot-cli run                     # play main scene
godot-cli run --scene res://levels/boss.tscn
godot-cli stop

godot-cli pause                   # toggle
godot-cli pause --on              # force pause
godot-cli pause --off             # force resume

godot-cli refresh                 # full filesystem scan
godot-cli refresh --sources       # source files only

godot-cli menu scene/save
godot-cli menu scene/save_all
godot-cli menu scene/reload
godot-cli menu filesystem/scan
```

`menu` is intentionally a whitelist; arbitrary editor menu paths are not invokable.

### Editor log

```bash
godot-cli log                                # last 50 lines
godot-cli log --lines 200 --type error
godot-cli log --type warn
godot-cli log --filter "ResourceLoader"
godot-cli log --clear                        # truncate after reading
```

The addon resolves the log file under `OS.get_user_data_dir()/logs/` and filters by `ERROR` / `WARNING` markers.

### exec — call a node method

```bash
godot-cli exec . get_name                    # method on the scene root
godot-cli exec Player get_hp
godot-cli exec UI/HUD update_score 42        # JSON-parseable args become typed
godot-cli exec . get_name --json
```

Node paths are relative to the currently edited scene's root. JSON-parseable arguments (`42`, `true`, `"foo"`, `{"k":1}`) are passed as typed values; everything else is treated as a string.

**Security:** `exec` is disabled by default. Enable it in **Editor → Editor Settings → `godot_cli/enable_exec` = true**.

### script — run a GDScript file

For multi-line logic where a single method call isn't enough:

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

Bare statements (no `extends`) are wrapped automatically: the addon prepends `extends Node` + `func _cli_execute():`, compiles via `GDScript.new().source_code`, attaches a transient `Node`, calls the entry method, and frees the node. Same `enable_exec` gate as `exec`.

### Scene tree

```bash
godot-cli node tree                          # default depth 3 from scene root
godot-cli node tree Player --depth 1 --props
godot-cli node get Player/Sprite scale
godot-cli node set Player/Sprite scale '{"x":2,"y":2}'
```

`node set` requires `enable_exec`.

### Resources

```bash
godot-cli resource find boss                          # path substring match
godot-cli resource info res://scenes/main.tscn

godot-cli reserialize --path res://scenes/main.tscn   # single file
godot-cli reserialize --all --dry-run                 # preview targets
godot-cli reserialize --all                           # commit
```

`reserialize` loads each `.tscn`/`.tres` through `ResourceLoader` and writes it back via `ResourceSaver`. Useful after text-editing scene/resource files in bulk.

### Project settings

```bash
godot-cli setting get application/config/name
godot-cli setting list                                # user-defined entries
godot-cli setting list --prefix myaddon/
godot-cli setting list --all                          # include built-ins
godot-cli setting set myaddon/threshold 0.5
```

`setting set` calls `ProjectSettings.save()` and modifies `project.godot` on disk; it requires `enable_exec`.

### Profiler

```bash
godot-cli profiler status
godot-cli profiler enable                             # start collecting on-demand samples
godot-cli profiler hierarchy                          # capture latest
godot-cli profiler hierarchy --frame 3 --top 20
godot-cli profiler clear
godot-cli profiler disable
```

Captures the `Performance` singleton: FPS, frame/physics time, static memory, object/node/orphan counts, draw calls.

### Screenshot

```bash
godot-cli screenshot --output /tmp/editor.png
godot-cli screenshot --output /tmp/game.png --target game
```

The addon captures the editor base control's viewport, encodes PNG, and streams base64 to the CLI which writes the file.

### Watch

```bash
godot-cli watch --once                                # single snapshot, then exit
godot-cli watch --path res://addons/ --interval 2     # poll loop with rescan on change
```

### Export

```bash
godot-cli export --preset "Windows Desktop" --output build/game.exe
godot-cli export --preset "Linux/X11"      --output build/game.x86_64 --debug
godot-cli export --preset Web              --output build/index.html --godot /opt/godot/godot
```

CLI-only — wraps `godot --headless --export-release` (or `--export-debug`). Auto-detects `godot` / `godot4` on `PATH`; override with `--godot <path>`.

### Tests (GUT)

```bash
godot-cli test                                # res://test/ by default
godot-cli test --directory res://tests/unit/
godot-cli test --filter test_player_
godot-cli test --mode editor                  # show GUT panel advice instead of spawning
godot-cli test --godot /opt/godot/godot
```

Requires the [GUT addon](https://github.com/bitwes/Gut) at `res://addons/gut/`. Standalone mode shells out to `godot --headless -s addons/gut/gut_cmdln.gd -gdir=<dir> -gexit`.

### Status, list, connect

```bash
godot-cli status
godot-cli connect
godot-cli list                                # asks the editor for registered tools
godot-cli list --local                        # CLI-side known commands only
```

## Global Options

| Flag | Description |
|------|-------------|
| `--port <n>` | Target a specific editor instance by port |
| `--project <path>` | Target a specific project directory |
| `--json` | Output raw JSON |

```bash
godot-cli --port 8091 status
godot-cli --project ~/games/mygame run
godot-cli status --json
```

Use `--help` on any command:

```bash
godot-cli node tree --help
godot-cli profiler hierarchy --help
```

## Multiple Editor Instances

Each running Godot editor binds its own port (8090, 8091, …) and writes its own instance file. The CLI matches by current working directory by default; otherwise the first arg with `--port` or `--project` wins.

```bash
ls ~/.godot-cli/instances/
# 12345.json  12390.json

godot-cli --port 8091 status
godot-cli --project ~/games/mygame run
godot-cli status            # cwd-match; single instance is auto-selected
```

The CLI removes instance files whose PID no longer exists before each command, so killed editors don't linger.

## Writing Custom Tools

Drop a `<name>_tool.gd` file into `addons/godot_cli/tools/`. The registry scans this directory at editor startup and registers it as command `<name>`. No registration call, no manifest, no rebuild.

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

Invoke it from anywhere:

```bash
curl -X POST http://127.0.0.1:8090/api/spawn \
  -H "Authorization: Bearer <token>" \
  -d '{"prefab":"res://prefabs/goblin.tscn","x":10,"y":5}'
```

**Conventions:**

- File name = command name. `spawn_tool.gd` → `spawn`. The `_tool.gd` suffix is required so the registry knows which scripts to load.
- Entry point: `func handle(request: Dictionary, _plugin: EditorPlugin) -> Dictionary`.
- The `request` dictionary contains `path`, `headers`, and `body` (raw JSON string). Parse `body` with `JSON.parse_string`.
- Responses use `ResponseBuilder.success(data)` or `ResponseBuilder.error(http_status, code, message, details?)`.
- For Godot types in the response (`Vector2`, `Color`, `Object`…), pass values through `util/variant_serializer.gd` so the client receives clean JSON.
- Auth is handled centrally — your tool only runs after the bearer token check passes.

To expose the tool as a typed Go subcommand, add a file under `cmd/` modeled on `cmd/run.go` (resolve via `cmdutil.Resolve`, send via `client.NewHTTPTransport`, render with `output.PrintJSON` when `--json` is set).

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Command error (see stderr) |
| 2 | No active Godot editor instance found |
| 3 | Authentication failed |

## Project Layout

```
godot-cli/
├── cmd/                          # cobra subcommands (one .go per command)
├── internal/
│   ├── client/                   # Transport interface + HTTP implementation
│   ├── cmdutil/                  # instance resolution helpers
│   ├── errors/                   # error codes + exit codes
│   ├── instance/                 # heartbeat + zombie cleanup
│   └── output/                   # JSON / text formatters
├── addons/godot_cli/
│   ├── plugin.cfg / plugin.gd    # EditorPlugin entry
│   ├── server/                   # TCPServer, request parser, auth, response builder
│   ├── instance/                 # instance file + heartbeat
│   ├── tools/                    # *_tool.gd — auto-discovered command handlers
│   └── util/                     # variant serializer
├── main.go
└── Makefile
```

## Building

```bash
make build                 # current platform binary
make snapshot              # all platforms via goreleaser
make release               # tag-based release
make clean                 # remove binaries and dist/
```

## License

MIT
