package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	clierrors "github.com/Sqoff/godot-cli/internal/errors"
)

var (
	flagExportPreset string
	flagExportOutput string
	flagExportDebug  bool
	flagExportGodot  string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Wrap `godot --headless --export-release/--export-debug`",
	RunE:  runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVar(&flagExportPreset, "preset", "", "export preset name (required)")
	exportCmd.Flags().StringVar(&flagExportOutput, "output", "", "output file path (required)")
	exportCmd.Flags().BoolVar(&flagExportDebug, "debug", false, "use --export-debug instead of --export-release")
	exportCmd.Flags().StringVar(&flagExportGodot, "godot", "", "godot binary path (default: 'godot' or 'godot4' on PATH)")
}

func runExport(_ *cobra.Command, _ []string) error {
	if flagExportPreset == "" {
		fmt.Fprintln(os.Stderr, "error: --preset is required")
		os.Exit(clierrors.ExitCommandError)
	}
	if flagExportOutput == "" {
		fmt.Fprintln(os.Stderr, "error: --output is required")
		os.Exit(clierrors.ExitCommandError)
	}
	if _, err := os.Stat("project.godot"); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "error: project.godot not found in current directory")
		os.Exit(clierrors.ExitCommandError)
	}

	godotBin := flagExportGodot
	if godotBin == "" {
		godotBin = locateGodot()
	}
	if godotBin == "" {
		fmt.Fprintln(os.Stderr, "error: cannot find godot binary on PATH; use --godot <path>")
		os.Exit(clierrors.ExitCommandError)
	}

	exportFlag := "--export-release"
	if flagExportDebug {
		exportFlag = "--export-debug"
	}

	c := exec.Command(godotBin, "--headless", exportFlag, flagExportPreset, flagExportOutput)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			os.Exit(ee.ExitCode())
		}
		fmt.Fprintln(os.Stderr, "export failed:", err)
		os.Exit(clierrors.ExitCommandError)
	}
	return nil
}

func locateGodot() string {
	candidates := []string{"godot", "godot4"}
	if runtime.GOOS == "windows" {
		candidates = []string{"godot.exe", "godot4.exe", "godot", "godot4"}
	}
	for _, name := range candidates {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return ""
}
