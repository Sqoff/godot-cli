package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/Sqoff/godot-cli/internal/client"
	"github.com/Sqoff/godot-cli/internal/cmdutil"
	clierrors "github.com/Sqoff/godot-cli/internal/errors"
)

var flagScene string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the game in the editor",
	RunE:  runRun,
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVar(&flagScene, "scene", "", "scene path to run (default: main scene)")
}

func runRun(_ *cobra.Command, _ []string) error {
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

	var params map[string]any
	if flagScene != "" {
		params = map[string]any{"scene": flagScene}
	}

	resp, err := cl.Send(ctx, "run", params)
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

	if flagScene != "" {
		fmt.Printf("Game started  scene=%s\n", flagScene)
	} else {
		fmt.Println("Game started")
	}
	return nil
}
