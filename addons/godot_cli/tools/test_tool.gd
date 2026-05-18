@tool
extends RefCounted

const ResponseBuilder := preload("res://addons/godot_cli/server/response_builder.gd")

func handle(request: Dictionary, _plugin: EditorPlugin) -> Dictionary:
	var body = JSON.parse_string(request.get("body", ""))
	if not body is Dictionary:
		body = {}
	var mode: String = body.get("mode", "standalone")
	var directory: String = body.get("directory", "")

	var gut_installed := DirAccess.dir_exists_absolute(ProjectSettings.globalize_path("res://addons/gut"))
	if not gut_installed:
		return ResponseBuilder.success({
			"installed": false,
			"message": "GUT addon not installed. Install via AssetLib or https://github.com/bitwes/Gut",
		})

	if directory.is_empty():
		if DirAccess.dir_exists_absolute(ProjectSettings.globalize_path("res://test")):
			directory = "res://test/"
		elif DirAccess.dir_exists_absolute(ProjectSettings.globalize_path("res://tests")):
			directory = "res://tests/"
		else:
			directory = "res://test/"

	var command := "godot --headless -s addons/gut/gut_cmdln.gd -gdir=%s -gexit" % directory

	match mode:
		"standalone":
			return ResponseBuilder.success({
				"installed": true,
				"mode": "standalone",
				"directory": directory,
				"command": command,
				"advice": "Run this command from the project root, or use 'godot-cli test --godot <path>'",
			})
		"editor":
			return ResponseBuilder.success({
				"installed": true,
				"mode": "editor",
				"directory": directory,
				"command": command,
				"message": "editor mode runs require GUT's panel; open the bottom GUT panel manually.",
			})
		_:
			return ResponseBuilder.error(400, "INVALID_PARAMS", "Unknown mode: %s" % mode)
