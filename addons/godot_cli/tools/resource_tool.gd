@tool
extends RefCounted

const ResponseBuilder := preload("res://addons/godot_cli/server/response_builder.gd")
const MAX_FIND := 1000

func handle(request: Dictionary, _plugin: EditorPlugin) -> Dictionary:
	var body = JSON.parse_string(request.get("body", ""))
	if not body is Dictionary:
		body = {}
	var action: String = body.get("action", "")
	match action:
		"find":
			return _find(body.get("pattern", ""))
		"info":
			return _info(body.get("path", ""))
		_:
			return ResponseBuilder.error(400, "INVALID_PARAMS", "Unknown action: %s" % action)

func _find(pattern: String) -> Dictionary:
	if pattern.is_empty():
		return ResponseBuilder.error(400, "INVALID_PARAMS", "Missing 'pattern'")
	var matches: Array = []
	_walk("res://", pattern, matches)
	return ResponseBuilder.success({"paths": matches, "count": matches.size()})

func _walk(dir_path: String, pattern: String, out: Array) -> void:
	if out.size() >= MAX_FIND:
		return
	var dir := DirAccess.open(dir_path)
	if dir == null:
		return
	dir.list_dir_begin()
	var fname := dir.get_next()
	while fname != "":
		if out.size() >= MAX_FIND:
			break
		var full := dir_path.path_join(fname)
		if dir.current_is_dir():
			if not fname.begins_with("."):
				_walk(full, pattern, out)
		else:
			if full.findn(pattern) >= 0:
				out.append(full)
		fname = dir.get_next()
	dir.list_dir_end()

func _info(path: String) -> Dictionary:
	if path.is_empty():
		return ResponseBuilder.error(400, "INVALID_PARAMS", "Missing 'path'")
	if not ResourceLoader.exists(path):
		return ResponseBuilder.error(404, "RESOURCE_NOT_FOUND", "Resource not found: %s" % path)
	var res := ResourceLoader.load(path)
	if res == null:
		return ResponseBuilder.error(500, "EDITOR_ERROR", "Failed to load: %s" % path)
	return ResponseBuilder.success({
		"path": path,
		"type": res.get_class(),
		"resource_path": res.resource_path,
	})
