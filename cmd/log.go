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
	flagLogLines  int
	flagLogType   string
	flagLogFilter string
	flagLogClear  bool
	flagLogFollow bool
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Read recent editor log lines",
	RunE:  runLog,
}

func init() {
	rootCmd.AddCommand(logCmd)
	logCmd.Flags().IntVar(&flagLogLines, "lines", 50, "number of lines to read (max 5000)")
	logCmd.Flags().StringVar(&flagLogType, "type", "all", "filter by type: all|error|warn|info")
	logCmd.Flags().StringVar(&flagLogFilter, "filter", "", "substring filter applied after type filter")
	logCmd.Flags().BoolVar(&flagLogClear, "clear", false, "truncate the log file after reading")
	logCmd.Flags().BoolVar(&flagLogFollow, "follow", false, "(stub) follow mode prints a notice; not yet streaming")
}

func runLog(_ *cobra.Command, _ []string) error {
	if flagLogFollow {
		fmt.Fprintln(os.Stderr, "warn: --follow not yet streaming; printing one snapshot")
	}
	inst, err := cmdutil.Resolve(flagPort, flagProject)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(clierrors.ExitNoConnection)
	}
	tr := client.NewHTTPTransport("127.0.0.1", inst.Port, inst.Token)
	defer tr.Close()
	cl := client.New(tr)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := cl.Send(ctx, "log", map[string]any{
		"lines":  flagLogLines,
		"type":   flagLogType,
		"filter": flagLogFilter,
		"clear":  flagLogClear,
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
		Lines []string `json:"lines"`
		Path  string   `json:"path"`
		Count int      `json:"count"`
	}
	_ = json.Unmarshal(resp.Data, &data)
	for _, line := range data.Lines {
		fmt.Println(line)
	}
	if data.Path != "" {
		fmt.Fprintf(os.Stderr, "(%d lines from %s)\n", data.Count, data.Path)
	}
	return nil
}
