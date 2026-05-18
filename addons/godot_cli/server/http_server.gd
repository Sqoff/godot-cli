@tool
extends Node

const Auth := preload("res://addons/godot_cli/server/auth.gd")
const RequestParser := preload("res://addons/godot_cli/server/request_parser.gd")
const ResponseBuilder := preload("res://addons/godot_cli/server/response_builder.gd")

const DEFAULT_PORT := 8090
const READ_TIMEOUT_MS := 2000

var _token: String = ""
var _tcp_server: TCPServer = null
var _pending: Array = []  # Array of {peer, data: PackedByteArray, started_at: int}
var _registry = null  # ToolRegistry

func set_token(token: String) -> void:
	_token = token

func set_registry(registry) -> void:
	_registry = registry

## Starts the server. Returns the bound port, or -1 on failure.
func start(port: int = DEFAULT_PORT) -> int:
	_tcp_server = TCPServer.new()

	var err := _tcp_server.listen(port)
	if err != OK:
		# Fallback: ask OS for any free port
		err = _tcp_server.listen(0)
	if err != OK:
		push_error("[godot-cli] TCPServer.listen failed: %d" % err)
		return -1

	return _tcp_server.get_local_port()

func stop() -> void:
	for conn in _pending:
		conn["peer"].disconnect_from_host()
	_pending.clear()

	if _tcp_server:
		_tcp_server.stop()
		_tcp_server = null

func _process(_delta: float) -> void:
	if _tcp_server == null or not _tcp_server.is_listening():
		return

	# Accept all waiting connections this frame
	while _tcp_server.is_connection_available():
		var peer: StreamPeerTCP = _tcp_server.take_connection()
		_pending.append({
			"peer": peer,
			"data": PackedByteArray(),
			"started_at": Time.get_ticks_msec()
		})

	# Advance each pending connection; remove finished ones
	var i := _pending.size() - 1
	while i >= 0:
		if _advance_connection(_pending[i]):
			_pending[i]["peer"].disconnect_from_host()
			_pending.remove_at(i)
		i -= 1

## Returns true when the connection should be closed (done or timed out).
func _advance_connection(conn: Dictionary) -> bool:
	var peer: StreamPeerTCP = conn["peer"]
	peer.poll()

	var available := peer.get_available_bytes()
	if available > 0:
		var read := peer.get_data(available)
		if read[0] == OK:
			conn["data"].append_array(read[1])

	var data: PackedByteArray = conn["data"]
	var text: String = data.get_string_from_utf8()
	var headers_end: int = text.find("\r\n\r\n")

	if headers_end != -1:
		var content_length := _parse_content_length(text.substr(0, headers_end))
		var body_received: int = data.size() - (headers_end + 4)
		if body_received >= content_length:
			_handle_request(peer, text)
			return true

	return Time.get_ticks_msec() - conn["started_at"] > READ_TIMEOUT_MS

func _handle_request(peer: StreamPeerTCP, raw: String) -> void:
	var request := RequestParser.parse(raw)
	var response: Dictionary

	if not Auth.validate(request.get("authorization", ""), _token):
		response = ResponseBuilder.error(401, "UNAUTHORIZED", "Invalid or missing token")
	else:
		response = _route(request)

	peer.put_data(ResponseBuilder.build_http_response(response))

func _route(request: Dictionary) -> Dictionary:
	if request.get("method", "") != "POST":
		return ResponseBuilder.error(405, "COMMAND_NOT_FOUND", "Only POST is supported")

	var path: String = request.get("path", "")

	if path == "/api/ping":
		return ResponseBuilder.success({"pong": true})

	if _registry == null:
		return ResponseBuilder.error(503, "NOT_READY", "Tool registry not initialized")

	return _registry.dispatch(request)

func _parse_content_length(headers_str: String) -> int:
	for line in headers_str.to_lower().split("\r\n"):
		if line.begins_with("content-length:"):
			var val := line.substr("content-length:".length()).strip_edges()
			if val.is_valid_int():
				return val.to_int()
	return 0
