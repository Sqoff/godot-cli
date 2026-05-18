@tool
extends RefCounted

static func validate(authorization_header: String, expected_token: String) -> bool:
	if not authorization_header.begins_with("Bearer "):
		return false
	var token := authorization_header.substr(7).strip_edges()
	return token == expected_token
