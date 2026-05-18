@tool
extends RefCounted

const ResponseBuilder    := preload("res://addons/godot_cli/server/response_builder.gd")
const VariantSerializer  := preload("res://addons/godot_cli/util/variant_serializer.gd")
const MAX_PROPS := 20

func handle(request: Dictionary, _plugin: EditorPlugin) -> Dictionary:
	var body = JSON.parse_string(request.get("body", ""))
	if not body is Dictionary:
		body = {}
	var action: String = body.get("action", "")
	var scene_root := EditorInterface.get_edited_scene_root()
	if scene_root == null:
		return ResponseBuilder.error(400, "NO_SCENE", "No scene is open in the editor")

	match action:
		"tree":
			return _action_tree(scene_root, body)
		"get":
			return _action_get(scene_root, body)
		"set":
			return _action_set(scene_root, body)
		_:
			return ResponseBuilder.error(400, "INVALID_PARAMS", "Unknown action: %s" % action)

func _action_tree(scene_root: Node, body: Dictionary) -> Dictionary:
	var path: String = body.get("path", "")
	var depth: int = int(body.get("depth", 3))
	var include_props: bool = bool(body.get("props", false))
	var origin: Node = scene_root if path.is_empty() else scene_root.get_node_or_null(path)
	if origin == null:
		return ResponseBuilder.error(404, "NODE_NOT_FOUND", "Node not found: %s" % path)
	return ResponseBuilder.success({"tree": _build_tree(origin, depth, include_props)})

func _build_tree(node: Node, depth: int, include_props: bool) -> Dictionary:
	var info := {
		"name": node.name,
		"type": node.get_class(),
		"path": str(node.get_path()),
		"children": [],
	}
	if include_props:
		info["properties"] = _collect_props(node)
	if depth > 0:
		for child in node.get_children():
			info["children"].append(_build_tree(child, depth - 1, include_props))
	return info

func _collect_props(node: Node) -> Dictionary:
	var out := {}
	var count := 0
	for prop in node.get_property_list():
		if count >= MAX_PROPS:
			break
		var usage: int = int(prop.get("usage", 0))
		if usage & PROPERTY_USAGE_EDITOR == 0:
			continue
		var pname: String = prop.get("name", "")
		if pname.is_empty() or pname.begins_with("_"):
			continue
		out[pname] = VariantSerializer.serialize(node.get(pname))
		count += 1
	return out

func _action_get(scene_root: Node, body: Dictionary) -> Dictionary:
	var path: String = body.get("path", "")
	var prop: String = body.get("property", "")
	if path.is_empty() or prop.is_empty():
		return ResponseBuilder.error(400, "INVALID_PARAMS", "Missing 'path' or 'property'")
	var node := scene_root.get_node_or_null(path)
	if node == null:
		return ResponseBuilder.error(404, "NODE_NOT_FOUND", "Node not found: %s" % path)
	var value = node.get(prop)
	return ResponseBuilder.success({"path": path, "property": prop, "value": VariantSerializer.serialize(value)})

func _action_set(scene_root: Node, body: Dictionary) -> Dictionary:
	var settings := EditorInterface.get_editor_settings()
	if not settings.get_setting("godot_cli/enable_exec"):
		return ResponseBuilder.error(403, "EXEC_DISABLED",
			"node set is disabled; enable via Editor > Editor Settings > godot_cli/enable_exec")
	var path: String = body.get("path", "")
	var prop: String = body.get("property", "")
	if path.is_empty() or prop.is_empty():
		return ResponseBuilder.error(400, "INVALID_PARAMS", "Missing 'path' or 'property'")
	var node := scene_root.get_node_or_null(path)
	if node == null:
		return ResponseBuilder.error(404, "NODE_NOT_FOUND", "Node not found: %s" % path)
	var value = body.get("value", null)
	node.set(prop, value)
	return ResponseBuilder.success({"ok": true, "path": path, "property": prop})
