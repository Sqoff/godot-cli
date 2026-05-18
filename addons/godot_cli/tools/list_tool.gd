@tool
extends RefCounted

const ResponseBuilder := preload("res://addons/godot_cli/server/response_builder.gd")
const TOOLS_DIR := "res://addons/godot_cli/tools/"
const TOOL_SUFFIX := "_tool.gd"

func handle(_request: Dictionary, _plugin: EditorPlugin) -> Dictionary:
	var commands: Array = []
	var dir := DirAccess.open(TOOLS_DIR)
	if dir == null:
		return ResponseBuilder.error(500, "INTERNAL_ERROR", "Failed to open tools directory")
	dir.list_dir_begin()
	var fname := dir.get_next()
	while fname != "":
		if not dir.current_is_dir() and fname.ends_with(TOOL_SUFFIX):
			commands.append(fname.substr(0, fname.length() - TOOL_SUFFIX.length()))
		fname = dir.get_next()
	dir.list_dir_end()
	commands.sort()
	return ResponseBuilder.success({"commands": commands})
