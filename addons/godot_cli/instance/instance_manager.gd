@tool
extends RefCounted

var _pid: int
var _port: int
var _token: String
var _instance_file: String
var _started_at: String

func setup(port: int, token: String) -> void:
	_pid = OS.get_process_id()
	_port = port
	_token = token
	_started_at = Time.get_datetime_string_from_system(true)
	_instance_file = _instances_dir() + "/%d.json" % _pid

func write_file() -> void:
	var dir := _instances_dir()
	if not DirAccess.dir_exists_absolute(dir):
		var err := DirAccess.make_dir_recursive_absolute(dir)
		if err != OK:
			push_error("[godot-cli] Cannot create dir: %s (err %d)" % [dir, err])
			return

	var data := {
		"pid": _pid,
		"port": _port,
		"project_path": ProjectSettings.globalize_path("res://"),
		"godot_version": _version_string(),
		"token": _token,
		"started_at": _started_at,
		"last_heartbeat": Time.get_datetime_string_from_system(true)
	}
	_atomic_write(_instance_file, JSON.stringify(data, "\t"))

func remove_file() -> void:
	if FileAccess.file_exists(_instance_file):
		var err := DirAccess.remove_absolute(_instance_file)
		if err != OK:
			push_error("[godot-cli] Cannot remove instance file (err %d)" % err)

func _atomic_write(path: String, content: String) -> void:
	var tmp := path + ".tmp"

	var f := FileAccess.open(tmp, FileAccess.WRITE)
	if not f:
		push_error("[godot-cli] Cannot write: %s (err %d)" % [tmp, FileAccess.get_open_error()])
		return
	f.store_string(content)
	f = null  # flush + close

	# rename within same directory — atomic on same filesystem (Unix + Windows)
	var dir := DirAccess.open(path.get_base_dir())
	if not dir:
		push_error("[godot-cli] Cannot open dir for rename: %s" % path.get_base_dir())
		DirAccess.remove_absolute(tmp)
		return

	var err := dir.rename(tmp.get_file(), path.get_file())
	if err != OK:
		push_error("[godot-cli] Rename failed (err %d)" % err)
		dir.remove(tmp.get_file())

func _instances_dir() -> String:
	var home := OS.get_environment("HOME")
	if home.is_empty():
		home = OS.get_environment("USERPROFILE")  # Windows fallback
	return home.rstrip("/\\") + "/.godot-cli/instances"

func _version_string() -> String:
	var v := Engine.get_version_info()
	return "%d.%d.%d" % [v.get("major", 0), v.get("minor", 0), v.get("patch", 0)]
