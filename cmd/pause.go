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
	flagPauseOn  bool
	flagPauseOff bool
)

var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause, resume, or toggle the running scene",
	RunE:  runPause,
}

func init() {
	rootCmd.AddCommand(pauseCmd)
	pauseCmd.Flags().BoolVar(&flagPauseOn, "on", false, "force pause")
	pauseCmd.Flags().BoolVar(&flagPauseOff, "off", false, "force resume")
}

func runPause(_ *cobra.Command, _ []string) error {
	if flagPauseOn && flagPauseOff {
		fmt.Fprintln(os.Stderr, "error: cannot use --on and --off together")
		os.Exit(clierrors.ExitCommandError)
	}
	action := "toggle"
	if flagPauseOn {
		action = "on"
	} else if flagPauseOff {
		action = "off"
	}

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

	resp, err := cl.Send(ctx, "pause", map[string]any{"action": action})
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
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		output.PrintJSON(resp.Data)
		return nil
	}
	if paused, _ := data["paused"].(bool); paused {
		fmt.Println("Paused")
	} else {
		fmt.Println("Resumed")
	}
	return nil
}
