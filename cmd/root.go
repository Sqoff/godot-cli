package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

var (
	flagPort    int
	flagProject string
	flagJSON    bool
)

var rootCmd = &cobra.Command{
	Use:     "godot-cli",
	Short:   "Godot Engine editor CLI controller",
	Long:    "Control Godot Engine editor from the command line.",
	Version: Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().IntVar(&flagPort, "port", 0, "override editor port")
	rootCmd.PersistentFlags().StringVar(&flagProject, "project", "", "Godot project path")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "output as JSON")
}
