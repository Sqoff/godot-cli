@tool
extends RefCounted

const ResponseBuilder := preload("res://addons/godot_cli/server/response_builder.gd")

func handle(_request: Dictionary, _plugin: EditorPlugin) -> Dictionary:
	EditorInterface.stop_playing_scene()
	return ResponseBuilder.success({"stopped": true})
