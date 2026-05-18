@tool
extends RefCounted

const ResponseBuilder := preload("res://addons/godot_cli/server/response_builder.gd")

const ALLOWED_ACTIONS := [
	"scene/save",
	"scene/save_all",
	"scene/reload",
	"filesystem/scan",
]

func handle(request: Dictionary, _plugin: EditorPlugin) -> Dictionary:
	var body = JSON.parse_string(request.get("body", ""))
	if not body is Dictionary:
		body = {}
	var action: String = body.get("action", "")
	if action.is_empty():
		return ResponseBuilder.error(400, "INVALID_PARAMS", "Missing 'action'")
	if not action in ALLOWED_ACTIONS:
		return ResponseBuilder.error(400, "UNKNOWN_ACTION",
			"Action not whitelisted: %s" % action,
			{"allowed": ALLOWED_ACTIONS})

	match action:
		"scene/save":
			EditorInterface.save_scene()
		"scene/save_all":
			EditorInterface.save_all_scenes()
		"scene/reload":
			var root = EditorInterface.get_edited_scene_root()
			if root == null:
				return ResponseBuilder.error(400, "NO_SCENE", "No scene is open")
			var path: String = root.scene_file_path
			if path.is_empty():
				return ResponseBuilder.error(400, "UNSAVED_SCENE", "Cannot reload an unsaved scene")
			EditorInterface.reload_scene_from_path(path)
		"filesystem/scan":
			var fs = EditorInterface.get_resource_filesystem()
			if fs == null:
				return ResponseBuilder.error(500, "EDITOR_ERROR", "Resource filesystem unavailable")
			fs.scan()

	return ResponseBuilder.success({"action": action, "ok": true})
