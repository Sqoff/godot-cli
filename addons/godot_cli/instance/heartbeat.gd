@tool
extends Node

const INTERVAL := 10.0  # seconds

var _manager = null  # InstanceManager
var _elapsed := 0.0

func setup(manager) -> void:
	_manager = manager

func _process(delta: float) -> void:
	_elapsed += delta
	if _elapsed >= INTERVAL:
		_elapsed = 0.0
		if _manager:
			_manager.write_file()
