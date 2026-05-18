@tool
extends RefCounted

const ResponseBuilder := preload("res://addons/godot_cli/server/response_builder.gd")

func handle(_request: Dictionary, _plugin: EditorPlugin) -> Dictionary:
	var scene_root := EditorInterface.get_edited_scene_root()
	var current_scene := scene_root.scene_file_path if scene_root else ""
	var v := Engine.get_version_info()
	var godot_version := "%d.%d.%d" % [v.get("major", 0), v.get("minor", 0), v.get("patch", 0)]
	return ResponseBuilder.success({
		"project":       ProjectSettings.globalize_path("res://"),
		"godot_version": godot_version,
		"current_scene": current_scene
	})
