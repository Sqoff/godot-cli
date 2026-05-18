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

var flagRefreshSources bool

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Re-scan the project asset filesystem",
	RunE:  runRefresh,
}

func init() {
	rootCmd.AddCommand(refreshCmd)
	refreshCmd.Flags().BoolVar(&flagRefreshSources, "sources", false, "scan only source files (faster, partial)")
}

func runRefresh(_ *cobra.Command, _ []string) error {
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

	resp, err := cl.Send(ctx, "refresh", map[string]any{"sources": flagRefreshSources})
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
	var data map[string]any
	_ = json.Unmarshal(resp.Data, &data)
	mode, _ := data["mode"].(string)
	if mode == "" {
		mode = "full"
	}
	fmt.Printf("Filesystem scan started (mode=%s)\n", mode)
	return nil
}
