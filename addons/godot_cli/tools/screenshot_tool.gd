@tool
extends RefCounted

const ResponseBuilder := preload("res://addons/godot_cli/server/response_builder.gd")

func handle(request: Dictionary, _plugin: EditorPlugin) -> Dictionary:
	var body = JSON.parse_string(request.get("body", ""))
	if not body is Dictionary:
		body = {}
	var target: String = body.get("target", "game")

	var viewport: Viewport = null
	var base := EditorInterface.get_base_control()
	if base:
		viewport = base.get_viewport()
	if viewport == null:
		return ResponseBuilder.error(500, "EDITOR_ERROR", "Could not access viewport")

	var tex := viewport.get_texture()
	if tex == null:
		return ResponseBuilder.error(500, "EDITOR_ERROR", "Viewport has no texture")
	var img := tex.get_image()
	if img == null:
		return ResponseBuilder.error(500, "EDITOR_ERROR", "Failed to convert viewport to image")

	var buf := img.save_png_to_buffer()
	if buf.is_empty():
		return ResponseBuilder.error(500, "EDITOR_ERROR", "PNG encode failed")

	return ResponseBuilder.success({
		"png_base64": Marshalls.raw_to_base64(buf),
		"width": img.get_width(),
		"height": img.get_height(),
		"target": target,
	})
