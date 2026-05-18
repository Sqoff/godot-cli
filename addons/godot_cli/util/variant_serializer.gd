@tool
extends RefCounted

static func serialize(value) -> Variant:
	var type := typeof(value)
	match type:
		TYPE_NIL:
			return null
		TYPE_BOOL, TYPE_INT, TYPE_FLOAT, TYPE_STRING:
			return value
		TYPE_VECTOR2:
			return {"x": value.x, "y": value.y}
		TYPE_VECTOR3:
			return {"x": value.x, "y": value.y, "z": value.z}
		TYPE_VECTOR4:
			return {"x": value.x, "y": value.y, "z": value.z, "w": value.w}
		TYPE_COLOR:
			return {"r": value.r, "g": value.g, "b": value.b, "a": value.a}
		TYPE_RECT2:
			return {"x": value.position.x, "y": value.position.y,
					"w": value.size.x, "h": value.size.y}
		TYPE_ARRAY:
			var out: Array = []
			for item in value:
				out.append(serialize(item))
			return out
		TYPE_DICTIONARY:
			var out: Dictionary = {}
			for key in value:
				out[str(key)] = serialize(value[key])
			return out
		TYPE_OBJECT:
			if value == null:
				return null
			return str(value)
		_:
			return str(value)
