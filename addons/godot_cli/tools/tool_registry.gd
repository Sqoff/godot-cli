@tool
extends RefCounted

const ResponseBuilder := preload("res://addons/godot_cli/server/response_builder.gd")

const TOOLS_DIR := "res://addons/godot_cli/tools/"
const TOOL_SUFFIX := "_tool.gd"

var _plugin: EditorPlugin = null
var _tools: Dictionary    = {}

func setup(plugin: EditorPlugin) -> void:
	_plugin = plugin
	_tools = {}
	_load_tools()

func _load_tools() -> void:
	var dir := DirAccess.open(TOOLS_DIR)
	if dir == null:
		push_error("[godot-cli] Failed to open tools directory: %s" % TOOLS_DIR)
		return
	dir.list_dir_begin()
	var fname := dir.get_next()
	while fname != "":
		if not dir.current_is_dir() and fname.ends_with(TOOL_SUFFIX):
			var command_name := fname.substr(0, fname.length() - TOOL_SUFFIX.length())
			_try_register(command_name, fname)
		fname = dir.get_next()
	dir.list_dir_end()

func _try_register(command_name: String, fname: String) -> bool:
	var script = load(TOOLS_DIR + fname)
	if script == null:
		push_warning("[godot-cli] load() returned null for %s" % fname)
		return false
	var inst = script.new()
	if inst == null:
		push_warning("[godot-cli] script.new() returned null for %s" % fname)
		return false
	_tools[command_name] = inst
	return true

func list_commands() -> Array:
	var commands := _tools.keys()
	commands.sort()
	return commands

func dispatch(request: Dictionary) -> Dictionary:
	var path: String = request.get("path", "")
	if not path.begins_with("/api/"):
		return ResponseBuilder.error(404, "COMMAND_NOT_FOUND", "Unknown path: %s" % path)

	var command := path.substr(5)

	# Lazy retry: if not loaded at setup time (e.g. EditorFileSystem race),
	# try to register it on first request.
	if not _tools.has(command):
		_try_register(command, command + TOOL_SUFFIX)

	if not _tools.has(command):
		return ResponseBuilder.error(404, "COMMAND_NOT_FOUND", "Unknown command: %s" % command)

	return _tools[command].handle(request, _plugin)
