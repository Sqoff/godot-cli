@tool
extends EditorPlugin

const HTTPServer       := preload("res://addons/godot_cli/server/http_server.gd")
const InstanceManager  := preload("res://addons/godot_cli/instance/instance_manager.gd")
const Heartbeat        := preload("res://addons/godot_cli/instance/heartbeat.gd")
const ToolRegistry     := preload("res://addons/godot_cli/tools/tool_registry.gd")

var _server: HTTPServer = null
var _heartbeat: Heartbeat = null
var _instance_manager        = null  # InstanceManager (RefCounted)
var _registry                = null  # ToolRegistry (RefCounted)

func _enter_tree() -> void:
	var token := _generate_token()

	# Step 1: HTTP server
	_server = HTTPServer.new()
	_server.set_token(token)
	add_child(_server)

	var port := _server.start()
	if port <= 0:
		push_error("[godot-cli] Failed to start HTTP server")
		return

	print("[godot-cli] Server running on port %d" % port)
	print("[godot-cli] Token: %s" % token)

	# Step 4: tool registry
	_registry = ToolRegistry.new()
	_registry.setup(self)
	_server.set_registry(_registry)

	# Step 2: editor settings
	_setup_editor_settings()

	# Step 2: instance file + heartbeat
	_instance_manager = InstanceManager.new()
	_instance_manager.setup(port, token)
	_instance_manager.write_file()

	_heartbeat = Heartbeat.new()
	_heartbeat.setup(_instance_manager)
	add_child(_heartbeat)

func _exit_tree() -> void:
	# stop heartbeat first so no write races cleanup
	if _heartbeat:
		_heartbeat.queue_free()
		_heartbeat = null

	if _instance_manager:
		_instance_manager.remove_file()
		_instance_manager = null

	if _server:
		_server.stop()
		_server.queue_free()
		_server = null

	_registry = null

func _generate_token() -> String:
	return Crypto.new().generate_random_bytes(16).hex_encode()

func _setup_editor_settings() -> void:
	var settings := EditorInterface.get_editor_settings()
	if not settings.has_setting("godot_cli/enable_exec"):
		settings.set_setting("godot_cli/enable_exec", false)
	settings.set_initial_value("godot_cli/enable_exec", false, false)
