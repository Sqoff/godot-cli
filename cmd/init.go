package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	clierrors "github.com/Sqoff/godot-cli/internal/errors"
)

var (
	flagInitForce  bool
	flagInitEnable bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Bootstrap the godot-cli plugin into the current Godot project",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&flagInitForce, "force", false, "overwrite existing plugin files")
	initCmd.Flags().BoolVar(&flagInitEnable, "enable", false, "also enable plugin in project.godot")
}

func runInit(_ *cobra.Command, _ []string) error {
	if _, err := os.Stat("project.godot"); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "error: project.godot not found; run this from your Godot project root")
		os.Exit(clierrors.ExitCommandError)
	}

	dest := filepath.Join("addons", "godot_cli")
	if _, err := os.Stat(dest); err == nil && !flagInitForce {
		fmt.Printf("Plugin already exists at %s\n", dest)
		fmt.Println("Use --force to overwrite")
	} else {
		if err := extractPlugin(dest); err != nil {
			fmt.Fprintf(os.Stderr, "install failed: %v\n", err)
			os.Exit(clierrors.ExitCommandError)
		}
		fmt.Printf("Plugin installed to ./%s\n", dest)
	}

	if flagInitEnable {
		if err := enablePluginInProject("project.godot"); err != nil {
			fmt.Fprintf(os.Stderr, "warn: failed to update project.godot: %v\n", err)
		} else {
			fmt.Println("Enabled plugin in project.godot")
		}
	} else {
		fmt.Println("Next: Project -> Project Settings -> Plugins -> Enable GodotCLI")
		fmt.Println("(or rerun with --enable to auto-enable)")
	}
	return nil
}

func enablePluginInProject(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := string(data)
	if strings.Contains(text, `"res://addons/godot_cli/plugin.cfg"`) {
		return nil
	}
	if strings.Contains(text, "[editor_plugins]") {
		return fmt.Errorf("project.godot already has [editor_plugins] section; add res://addons/godot_cli/plugin.cfg to it manually")
	}
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	text += "\n[editor_plugins]\nenabled=PackedStringArray(\"res://addons/godot_cli/plugin.cfg\")\n"
	return os.WriteFile(path, []byte(text), 0644)
}
