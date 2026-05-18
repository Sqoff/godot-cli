package main

import (
	"embed"

	"github.com/Sqoff/godot-cli/cmd"
)

//go:embed addons/godot_cli
var pluginFS embed.FS

func main() {
	cmd.PluginFS = pluginFS
	cmd.Execute()
}
