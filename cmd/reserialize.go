package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/Sqoff/godot-cli/internal/client"
	"github.com/Sqoff/godot-cli/internal/cmdutil"
	clierrors "github.com/Sqoff/godot-cli/internal/errors"
	"github.com/Sqoff/godot-cli/internal/output"
)

var (
	flagReserializePath   string
	flagReserializeAll    bool
	flagReserializeDryRun bool
)

var reserializeCmd = &cobra.Command{
	Use:   "reserialize",
	Short: "Re-save .tscn/.tres resources through ResourceLoader/Saver",
	RunE:  runReserialize,
}

func init() {
	rootCmd.AddCommand(reserializeCmd)
	reserializeCmd.Flags().StringVar(&flagReserializePath, "path", "", "single resource path (res://...)")
	reserializeCmd.Flags().BoolVar(&flagReserializeAll, "all", false, "process all .tscn/.tres in project")
	reserializeCmd.Flags().BoolVar(&flagReserializeDryRun, "dry-run", false, "list targets without saving")
}

func runReserialize(_ *cobra.Command, _ []string) error {
	if flagReserializePath == "" && !flagReserializeAll {
		fmt.Fprintln(os.Stderr, "error: provide --path <res://...> or --all")
		os.Exit(clierrors.ExitCommandError)
	}

	inst, err := cmdutil.Resolve(flagPort, flagProject)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(clierrors.ExitNoConnection)
	}
	tr := client.NewHTTPTransport("127.0.0.1", inst.Port, inst.Token)
	defer tr.Close()
	cl := client.New(tr)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	resp, err := cl.Send(ctx, "reserialize", map[string]any{
		"path":    flagReserializePath,
		"all":     flagReserializeAll,
		"dry_run": flagReserializeDryRun,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "connection error:", err)
		os.Exit(clierrors.ExitNoConnection)
	}
	if !resp.Success {
		if resp.Error != nil {
			fmt.Fprintln(os.Stderr, "error:", resp.Error.Message)
		}
		os.Exit(clierrors.ExitCommandError)
	}
	if flagJSON {
		output.PrintJSON(resp.Data)
		return nil
	}
	var data struct {
		Processed []string `json:"processed"`
		Count     int      `json:"count"`
		DryRun    bool     `json:"dry_run"`
	}
	_ = json.Unmarshal(resp.Data, &data)
	if data.DryRun {
		fmt.Printf("Dry run: %d resource(s) would be re-saved\n", data.Count)
	} else {
		fmt.Printf("Re-saved %d resource(s)\n", data.Count)
	}
	for _, p := range data.Processed {
		fmt.Printf("  %s\n", p)
	}
	return nil
}
