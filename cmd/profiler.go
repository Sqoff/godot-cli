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

var profilerCmd = &cobra.Command{
	Use:   "profiler",
	Short: "Capture and inspect editor performance monitors",
}

var profilerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show profiler sampling state",
	RunE:  runProfilerAction("status"),
}
var profilerEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Start capturing samples on demand",
	RunE:  runProfilerAction("enable"),
}
var profilerDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Stop capturing samples",
	RunE:  runProfilerAction("disable"),
}
var profilerClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Discard captured samples",
	RunE:  runProfilerAction("clear"),
}

var (
	flagProfilerFrame int
	flagProfilerTop   int
)

var profilerHierarchyCmd = &cobra.Command{
	Use:   "hierarchy",
	Short: "Capture and return current performance monitor snapshot",
	RunE:  runProfilerAction("hierarchy"),
}

func init() {
	rootCmd.AddCommand(profilerCmd)
	profilerCmd.AddCommand(profilerStatusCmd)
	profilerCmd.AddCommand(profilerEnableCmd)
	profilerCmd.AddCommand(profilerDisableCmd)
	profilerCmd.AddCommand(profilerClearCmd)
	profilerCmd.AddCommand(profilerHierarchyCmd)
	profilerHierarchyCmd.Flags().IntVar(&flagProfilerFrame, "frame", -1, "frame index (negative = latest capture)")
	profilerHierarchyCmd.Flags().IntVar(&flagProfilerTop, "top", 10, "max monitor entries to return")
}

func runProfilerAction(action string) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		inst, err := cmdutil.Resolve(flagPort, flagProject)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(clierrors.ExitNoConnection)
		}
		tr := client.NewHTTPTransport("127.0.0.1", inst.Port, inst.Token)
		defer tr.Close()
		cl := client.New(tr)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		params := map[string]any{"action": action}
		if action == "hierarchy" {
			params["frame"] = flagProfilerFrame
			params["top"] = flagProfilerTop
		}
		resp, err := cl.Send(ctx, "profiler", params)
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
		var generic map[string]any
		_ = json.Unmarshal(resp.Data, &generic)
		out, _ := json.MarshalIndent(generic, "", "  ")
		fmt.Println(string(out))
		return nil
	}
}
