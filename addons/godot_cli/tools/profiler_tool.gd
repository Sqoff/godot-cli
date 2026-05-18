@tool
extends RefCounted

const ResponseBuilder := preload("res://addons/godot_cli/server/response_builder.gd")
const MAX_SAMPLES := 600

var _enabled: bool = false
var _samples: Array = []

func handle(request: Dictionary, _plugin: EditorPlugin) -> Dictionary:
	var body = JSON.parse_string(request.get("body", ""))
	if not body is Dictionary:
		body = {}
	var action: String = body.get("action", "")
	match action:
		"status":
			return ResponseBuilder.success({
				"enabled": _enabled,
				"sample_count": _samples.size(),
			})
		"enable":
			_enabled = true
			return ResponseBuilder.success({"enabled": true})
		"disable":
			_enabled = false
			return ResponseBuilder.success({"enabled": false})
		"clear":
			_samples = []
			return ResponseBuilder.success({"cleared": true, "sample_count": 0})
		"hierarchy":
			var snap := _capture()
			if _enabled and _samples.size() < MAX_SAMPLES:
				_samples.append(snap)
			var frame_idx: int = int(body.get("frame", -1))
			var top: int = int(body.get("top", 10))
			var chosen := snap
			if frame_idx >= 0 and frame_idx < _samples.size():
				chosen = _samples[frame_idx]
			var resolved_frame := frame_idx
			if resolved_frame < 0:
				resolved_frame = max(_samples.size() - 1, 0)
			return ResponseBuilder.success({
				"frame": resolved_frame,
				"monitors": chosen,
				"top": top,
			})
		_:
			return ResponseBuilder.error(400, "INVALID_PARAMS", "Unknown action: %s" % action)

func _capture() -> Dictionary:
	return {
		"fps": Performance.get_monitor(Performance.TIME_FPS),
		"frame_time_ms": Performance.get_monitor(Performance.TIME_PROCESS) * 1000.0,
		"physics_time_ms": Performance.get_monitor(Performance.TIME_PHYSICS_PROCESS) * 1000.0,
		"memory_static_mb": Performance.get_monitor(Performance.MEMORY_STATIC) / (1024.0 * 1024.0),
		"memory_static_max_mb": Performance.get_monitor(Performance.MEMORY_STATIC_MAX) / (1024.0 * 1024.0),
		"objects": int(Performance.get_monitor(Performance.OBJECT_COUNT)),
		"nodes": int(Performance.get_monitor(Performance.OBJECT_NODE_COUNT)),
		"orphan_nodes": int(Performance.get_monitor(Performance.OBJECT_ORPHAN_NODE_COUNT)),
		"draw_calls": int(Performance.get_monitor(Performance.RENDER_TOTAL_DRAW_CALLS_IN_FRAME)),
	}
