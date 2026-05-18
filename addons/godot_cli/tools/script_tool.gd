@tool
extends RefCounted

const ResponseBuilder    := preload("res://addons/godot_cli/server/response_builder.gd")
const VariantSerializer  := preload("res://addons/godot_cli/util/variant_serializer.gd")

func handle(request: Dictionary, _plugin: EditorPlugin) -> Dictionary:
	var settings := EditorInterface.get_editor_settings()
	if not settings.get_setting("godot_cli/enable_exec"):
		return ResponseBuilder.error(403, "EXEC_DISABLED",
			"script is disabled; enable via Editor > Editor Settings > godot_cli/enable_exec")

	var body = JSON.parse_string(request.get("body", ""))
	if not body is Dictionary:
		body = {}
	var code: String = body.get("code", "")
	if code.strip_edges().is_empty():
		return ResponseBuilder.error(400, "INVALID_PARAMS", "Missing 'code'")
	var method_name: String = body.get("method", "_cli_execute")

	var wrapped := _wrap_if_needed(code, method_name)

	var script := GDScript.new()
	script.source_code = wrapped
	var err := script.reload()
	if err != OK:
		return ResponseBuilder.error(400, "INVALID_PARAMS",
			"GDScript compile failed", {"error_code": err})

	var runner := Node.new()
	runner.name = "_GodotCliScriptRunner"
	runner.set_script(script)
	EditorInterface.get_base_control().add_child(runner)

	var result = null
	if runner.has_method(method_name):
		result = runner.call(method_name)

	runner.queue_free()

	return ResponseBuilder.success({"result": VariantSerializer.serialize(result)})

func _wrap_if_needed(code: String, method_name: String) -> String:
	if code.find("extends ") >= 0:
		return code
	var indented := ""
	for line in code.split("\n"):
		indented += "\t" + line + "\n"
	return "extends Node\nfunc %s():\n%s" % [method_name, indented]
