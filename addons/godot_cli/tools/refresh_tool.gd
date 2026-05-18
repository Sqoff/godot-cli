@tool
extends RefCounted

const ResponseBuilder := preload("res://addons/godot_cli/server/response_builder.gd")

func handle(request: Dictionary, _plugin: EditorPlugin) -> Dictionary:
	var body = JSON.parse_string(request.get("body", ""))
	if not body is Dictionary:
		body = {}
	var sources_only: bool = bool(body.get("sources", false))

	var fs = EditorInterface.get_resource_filesystem()
	if fs == null:
		return ResponseBuilder.error(500, "EDITOR_ERROR", "Resource filesystem unavailable")

	if sources_only:
		fs.scan_sources()
	else:
		fs.scan()

	var mode := "sources" if sources_only else "full"
	return ResponseBuilder.success({"scan_started": true, "mode": mode})
