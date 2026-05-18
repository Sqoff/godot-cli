@tool
extends RefCounted

const ResponseBuilder := preload("res://addons/godot_cli/server/response_builder.gd")
const MAX_LINES := 5000

func handle(request: Dictionary, _plugin: EditorPlugin) -> Dictionary:
	var body = JSON.parse_string(request.get("body", ""))
	if not body is Dictionary:
		body = {}
	var n: int = clampi(int(body.get("lines", 50)), 1, MAX_LINES)
	var type_filter: String = body.get("type", "all")
	var substr: String = body.get("filter", "")
	var should_clear: bool = bool(body.get("clear", false))

	var path := _find_log_path()
	if path.is_empty():
		return ResponseBuilder.error(404, "LOG_NOT_FOUND",
			"No godot.log file located under user data dir")

	var lines := _read_tail(path, n * 4)
	lines = _filter_type(lines, type_filter)
	if not substr.is_empty():
		lines = _filter_substr(lines, substr)

	if lines.size() > n:
		lines = lines.slice(lines.size() - n, lines.size())

	if should_clear:
		var f := FileAccess.open(path, FileAccess.WRITE)
		if f:
			f.close()

	return ResponseBuilder.success({
		"lines": lines,
		"path": ProjectSettings.globalize_path(path),
		"count": lines.size(),
	})

func _find_log_path() -> String:
	var user_dir := OS.get_user_data_dir()
	var candidates := [
		user_dir.path_join("logs/godot.log"),
		user_dir.path_join("../logs/godot.log"),
	]
	for c in candidates:
		if FileAccess.file_exists(c):
			return c
	var d := DirAccess.open(user_dir.path_join("logs"))
	if d:
		d.list_dir_begin()
		var name := d.get_next()
		while name != "":
			if name.ends_with(".log"):
				d.list_dir_end()
				return user_dir.path_join("logs").path_join(name)
			name = d.get_next()
		d.list_dir_end()
	return ""

func _read_tail(path: String, max_lines: int) -> Array:
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		return []
	var content := f.get_as_text()
	f.close()
	var split := content.split("\n", false)
	var arr: Array = []
	var start: int = max(0, split.size() - max_lines)
	for i in range(start, split.size()):
		arr.append(split[i])
	return arr

func _filter_type(lines: Array, type_filter: String) -> Array:
	if type_filter == "" or type_filter == "all":
		return lines
	match type_filter:
		"error":
			var out: Array = []
			for ln in lines:
				if ln.findn("ERROR") >= 0:
					out.append(ln)
			return out
		"warn":
			var out: Array = []
			for ln in lines:
				if ln.findn("WARNING") >= 0:
					out.append(ln)
			return out
		"info":
			var out: Array = []
			for ln in lines:
				if ln.findn("ERROR") < 0 and ln.findn("WARNING") < 0:
					out.append(ln)
			return out
	return lines

func _filter_substr(lines: Array, pattern: String) -> Array:
	var out: Array = []
	for ln in lines:
		if ln.findn(pattern) >= 0:
			out.append(ln)
	return out
