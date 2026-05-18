@tool
extends RefCounted

const ResponseBuilder := preload("res://addons/godot_cli/server/response_builder.gd")
const MAX_FILES := 2000

func handle(request: Dictionary, _plugin: EditorPlugin) -> Dictionary:
	var body = JSON.parse_string(request.get("body", ""))
	if not body is Dictionary:
		body = {}
	var action: String = body.get("action", "snapshot")
	var path: String = body.get("path", "res://")
	if path.is_empty():
		path = "res://"

	match action:
		"snapshot":
			var files: Array = []
			_walk(path, files)
			return ResponseBuilder.success({"files": files, "count": files.size(), "path": path})
		"scan":
			var fs = EditorInterface.get_resource_filesystem()
			if fs == null:
				return ResponseBuilder.error(500, "EDITOR_ERROR", "Resource filesystem unavailable")
			fs.scan()
			return ResponseBuilder.success({"scan_started": true, "path": path})
		_:
			return ResponseBuilder.error(400, "INVALID_PARAMS", "Unknown action: %s" % action)

func _walk(dir_path: String, out: Array) -> void:
	if out.size() >= MAX_FILES:
		return
	var dir := DirAccess.open(dir_path)
	if dir == null:
		return
	dir.list_dir_begin()
	var fname := dir.get_next()
	while fname != "":
		if out.size() >= MAX_FILES:
			break
		var full := dir_path.path_join(fname)
		if dir.current_is_dir():
			if not fname.begins_with("."):
				_walk(full, out)
		else:
			var mtime := FileAccess.get_modified_time(full)
			out.append({"path": full, "mtime": mtime})
		fname = dir.get_next()
	dir.list_dir_end()
