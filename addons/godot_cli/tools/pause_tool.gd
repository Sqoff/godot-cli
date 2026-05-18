@tool
extends RefCounted

const ResponseBuilder := preload("res://addons/godot_cli/server/response_builder.gd")

func handle(request: Dictionary, _plugin: EditorPlugin) -> Dictionary:
	var body = JSON.parse_string(request.get("body", ""))
	if not (body is Dictionary):
		body = {}
	var action: String = body.get("action", "toggle")

	# Godot 4.x does not expose a stable pause/resume API on EditorInterface.
	# Probe dynamically via has_method to avoid parse-time errors on engine
	# builds where the method is absent, and degrade to NOT_IMPLEMENTED.
	if not EditorInterface.has_method("set_pause_scene"):
		return ResponseBuilder.error(501, "NOT_IMPLEMENTED",
			"EditorInterface has no public pause API in this Godot version; control pause via the running scene's debugger UI")

	var current_paused: bool = false
	if EditorInterface.has_method("is_paused_scene"):
		current_paused = EditorInterface.call("is_paused_scene")

	var target: bool = current_paused
	match action:
		"on":
			target = true
		"off":
			target = false
		"toggle":
			target = not current_paused
		_:
			return ResponseBuilder.error(400, "INVALID_PARAMS", "Unknown action: %s" % action)

	EditorInterface.call("set_pause_scene", target)
	return ResponseBuilder.success({"paused": target})
