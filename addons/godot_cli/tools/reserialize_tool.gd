@tool
extends RefCounted

const ResponseBuilder := preload("res://addons/godot_cli/server/response_builder.gd")
const MAX_FILES := 5000

func handle(request: Dictionary, _plugin: EditorPlugin) -> Dictionary:
	var body = JSON.parse_string(request.get("body", ""))
	if not body is Dictionary:
		body = {}
	var single_path: String = body.get("path", "")
	var all_flag: bool = bool(body.get("all", false))
	var dry_run: bool = bool(body.get("dry_run", false))

	var targets: Array = []
	if not single_path.is_empty():
		targets.append(single_path)
	elif all_flag:
		_collect("res://", targets)
	else:
		return ResponseBuilder.error(400, "INVALID_PARAMS", "Provide 'path' or 'all'")

	var processed: Array = []
	for path in targets:
		if processed.size() >= MAX_FILES:
			break
		if dry_run:
			processed.append(path)
			continue
		var res = ResourceLoader.load(path)
		if res == null:
			continue
		var err := ResourceSaver.save(res, path)
		if err == OK:
			processed.append(path)

	return ResponseBuilder.success({
		"processed": processed,
		"count": processed.size(),
		"dry_run": dry_run,
	})

func _collect(dir_path: String, out: Array) -> void:
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
				_collect(full, out)
		else:
			if fname.ends_with(".tscn") or fname.ends_with(".tres"):
				out.append(full)
		fname = dir.get_next()
	dir.list_dir_end()
