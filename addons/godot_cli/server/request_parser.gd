@tool
extends RefCounted

static func parse(raw: String) -> Dictionary:
	var result := {
		"method": "",
		"path": "",
		"headers": {},
		"body": "",
		"authorization": ""
	}

	var parts := raw.split("\r\n\r\n", true, 1)
	if parts.is_empty():
		return result

	var header_section := parts[0]
	if parts.size() > 1:
		result["body"] = parts[1]

	var lines := header_section.split("\r\n")
	if lines.is_empty():
		return result

	# Parse request line: METHOD /path HTTP/1.1
	var request_line := lines[0].split(" ")
	if request_line.size() >= 2:
		result["method"] = request_line[0].to_upper()
		result["path"] = request_line[1]

	# Parse headers
	for i in range(1, lines.size()):
		var line := lines[i]
		var colon_idx := line.find(":")
		if colon_idx == -1:
			continue
		var key := line.substr(0, colon_idx).strip_edges().to_lower()
		var value := line.substr(colon_idx + 1).strip_edges()
		result["headers"][key] = value
		if key == "authorization":
			result["authorization"] = value

	return result
