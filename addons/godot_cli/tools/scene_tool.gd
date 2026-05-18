@tool
extends RefCounted

const ResponseBuilder := preload("res://addons/godot_cli/server/response_builder.gd")

func handle(request: Dictionary, _plugin: EditorPlugin) -> Dictionary:
	var body: Dictionary = _parse_body(request.get("body", ""))
	var action: String = body.get("action", "get")

	match action:
		"get":
			var scene_root := EditorInterface.get_edited_scene_root()
			var path := scene_root.scene_file_path if scene_root else ""
			return ResponseBuilder.success({"scene": path})
		"set":
			var path: String = body.get("path", "")
			if path.is_empty():
				return ResponseBuilder.error(400, "INVALID_PARAMS", "Missing 'path' field")
			EditorInterface.open_scene_from_path(path)
			return ResponseBuilder.success({"scene": path})
		_:
			return ResponseBuilder.error(400, "INVALID_PARAMS", "Unknown action: %s" % action)

func _parse_body(body_str: String) -> Dictionary:
	if body_str.is_empty():
		return {}
	var parsed = JSON.parse_string(body_str)
	return parsed if parsed is Dictionary else {}
