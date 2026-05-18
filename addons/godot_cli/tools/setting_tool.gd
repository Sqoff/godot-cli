@tool
extends RefCounted

const ResponseBuilder    := preload("res://addons/godot_cli/server/response_builder.gd")
const VariantSerializer  := preload("res://addons/godot_cli/util/variant_serializer.gd")
const MAX_LIST := 500
const BUILTIN_PREFIXES := [
	"application/",
	"rendering/",
	"physics/",
	"audio/",
	"input/",
	"internationalization/",
	"display/",
	"network/",
	"debug/",
	"gui/",
	"layer_names/",
	"shader_globals/",
	"editor_plugins/",
]

func handle(request: Dictionary, _plugin: EditorPlugin) -> Dictionary:
	var body = JSON.parse_string(request.get("body", ""))
	if not (body is Dictionary):
		body = {}
	var action: String = body.get("action", "")
	match action:
		"get":
			return _do_get(body.get("key", ""))
		"set":
			return _do_set(body.get("key", ""), body.get("value", null))
		"list":
			return _do_list(body.get("prefix", ""), bool(body.get("all", false)))
		_:
			return ResponseBuilder.error(400, "INVALID_PARAMS", "Unknown action: %s" % action)

func _do_get(key: String) -> Dictionary:
	if key.is_empty():
		return ResponseBuilder.error(400, "INVALID_PARAMS", "Missing 'key'")
	if not ProjectSettings.has_setting(key):
		return ResponseBuilder.error(404, "SETTING_NOT_FOUND", "Setting not found: %s" % key)
	return ResponseBuilder.success({
		"key": key,
		"value": VariantSerializer.serialize(ProjectSettings.get_setting(key)),
	})

func _do_set(key: String, value) -> Dictionary:
	var settings := EditorInterface.get_editor_settings()
	if not settings.get_setting("godot_cli/enable_exec"):
		return ResponseBuilder.error(403, "EXEC_DISABLED",
			"setting set is disabled; enable via Editor > Editor Settings > godot_cli/enable_exec")
	if key.is_empty():
		return ResponseBuilder.error(400, "INVALID_PARAMS", "Missing 'key'")
	ProjectSettings.set_setting(key, value)
	var err := ProjectSettings.save()
	if err != OK:
		return ResponseBuilder.error(500, "EDITOR_ERROR", "ProjectSettings.save() failed", {"error_code": err})
	return ResponseBuilder.success({"key": key, "ok": true})

func _do_list(prefix: String, include_builtin: bool) -> Dictionary:
	var items: Array = []
	for prop in ProjectSettings.get_property_list():
		if items.size() >= MAX_LIST:
			break
		var pname: String = prop.get("name", "")
		if pname.is_empty():
			continue
		if not prefix.is_empty() and not pname.begins_with(prefix):
			continue
		if not include_builtin and _is_builtin(pname):
			continue
		items.append({
			"name": pname,
			"value": VariantSerializer.serialize(ProjectSettings.get_setting(pname)),
		})
	return ResponseBuilder.success({"settings": items, "count": items.size()})

func _is_builtin(pname: String) -> bool:
	for prefix in BUILTIN_PREFIXES:
		if pname.begins_with(prefix):
			return true
	return false
