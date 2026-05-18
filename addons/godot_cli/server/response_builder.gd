@tool
extends RefCounted

static func success(data: Variant = null) -> Dictionary:
	return {"http_status": 200, "success": true, "data": data}

static func error(http_status: int, code: String, message: String, details: Dictionary = {}) -> Dictionary:
	return {
		"http_status": http_status,
		"success": false,
		"error": {"code": code, "message": message, "details": details}
	}

static func build_http_response(response: Dictionary) -> PackedByteArray:
	var status: int = response.get("http_status", 200)

	# Build JSON body excluding the internal http_status field
	var json_body := {}
	for key in response:
		if key != "http_status":
			json_body[key] = response[key]

	var body_bytes := JSON.stringify(json_body).to_utf8_buffer()

	var headers := (
		"HTTP/1.1 %d %s\r\n" % [status, _status_text(status)]
		+ "Content-Type: application/json\r\n"
		+ "Content-Length: %d\r\n" % body_bytes.size()
		+ "Connection: close\r\n"
		+ "\r\n"
	)

	var result := PackedByteArray()
	result.append_array(headers.to_utf8_buffer())
	result.append_array(body_bytes)
	return result

static func _status_text(code: int) -> String:
	match code:
		200: return "OK"
		400: return "Bad Request"
		401: return "Unauthorized"
		403: return "Forbidden"
		404: return "Not Found"
		405: return "Method Not Allowed"
		408: return "Request Timeout"
		500: return "Internal Server Error"
	return "Unknown"
