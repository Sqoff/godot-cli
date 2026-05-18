package cmd

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var PluginFS embed.FS

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage the Godot plugin",
}

var pluginInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the Godot plugin into the current project",
	RunE:  runPluginInstall,
}

var flagForce bool

func init() {
	rootCmd.AddCommand(pluginCmd)
	pluginCmd.AddCommand(pluginInstallCmd)
	pluginInstallCmd.Flags().BoolVar(&flagForce, "force", false, "overwrite existing plugin files")
}

func runPluginInstall(_ *cobra.Command, _ []string) error {
	if _, err := os.Stat("project.godot"); os.IsNotExist(err) {
		return fmt.Errorf("project.godot not found; run this command from your Godot project root")
	}

	dest := filepath.Join("addons", "godot_cli")
	if _, err := os.Stat(dest); err == nil && !flagForce {
		fmt.Printf("Plugin already exists at %s\n", dest)
		fmt.Println("Use --force to overwrite")
		return nil
	}

	if err := extractPlugin(dest); err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	fmt.Printf("Plugin installed to ./%s\n", dest)
	fmt.Println("Next: Project -> Project Settings -> Plugins -> Enable GodotCLI")
	return nil
}

func extractPlugin(dest string) error {
	return fs.WalkDir(PluginFS, "addons/godot_cli", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel("addons/godot_cli", path)
		target := filepath.Join(dest, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		return copyEmbedFile(path, target)
	})
}

func copyEmbedFile(src, dest string) error {
	in, err := PluginFS.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
