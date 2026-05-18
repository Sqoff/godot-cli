@tool
extends RefCounted

const ResponseBuilder    := preload("res://addons/godot_cli/server/response_builder.gd")
const VariantSerializer  := preload("res://addons/godot_cli/util/variant_serializer.gd")

func handle(request: Dictionary, _plugin: EditorPlugin) -> Dictionary:
	var settings := EditorInterface.get_editor_settings()
	if not settings.get_setting("godot_cli/enable_exec"):
		return ResponseBuilder.error(403, "EXEC_DISABLED",
			"exec is disabled; enable via Editor > Editor Settings > godot_cli/enable_exec")

	var body: Dictionary = _parse_body(request.get("body", ""))
	if body.is_empty():
		return ResponseBuilder.error(400, "INVALID_PARAMS", "Request body is required")

	var node_path: String = body.get("node", "")
	var method: String    = body.get("method", "")
	var args: Array       = body.get("args", [])

	if node_path.is_empty():
		return ResponseBuilder.error(400, "INVALID_PARAMS", "Missing 'node' field")
	if method.is_empty():
		return ResponseBuilder.error(400, "INVALID_PARAMS", "Missing 'method' field")

	var scene_root := EditorInterface.get_edited_scene_root()
	if scene_root == null:
		return ResponseBuilder.error(400, "NO_SCENE", "No scene is open in the editor")

	var target := scene_root.get_node_or_null(node_path)
	if target == null:
		return ResponseBuilder.error(404, "NODE_NOT_FOUND", "Node not found: %s" % node_path)

	if not target.has_method(method):
		return ResponseBuilder.error(404, "METHOD_NOT_FOUND",
			"Method '%s' not found on node '%s'" % [method, node_path])

	var result = target.callv(method, args)
	return ResponseBuilder.success({"result": VariantSerializer.serialize(result)})

func _parse_body(body_str: String) -> Dictionary:
	if body_str.is_empty():
		return {}
	var parsed = JSON.parse_string(body_str)
	return parsed if parsed is Dictionary else {}
