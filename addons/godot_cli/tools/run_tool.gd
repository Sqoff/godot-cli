@tool
extends RefCounted

const ResponseBuilder := preload("res://addons/godot_cli/server/response_builder.gd")

func handle(request: Dictionary, _plugin: EditorPlugin) -> Dictionary:
	var scene := _parse_scene(request.get("body", ""))
	if scene != "":
		EditorInterface.open_scene_from_path(scene)
		EditorInterface.play_current_scene()
	else:
		EditorInterface.play_main_scene()
	return ResponseBuilder.success({"started": true})

func _parse_scene(body_str: String) -> String:
	if body_str.is_empty():
		return ""
	var body = JSON.parse_string(body_str)
	if body is Dictionary:
		return body.get("scene", "")
	return ""
