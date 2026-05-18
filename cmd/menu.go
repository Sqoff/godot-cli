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

var menuCmd = &cobra.Command{
	Use:   "menu <action>",
	Short: "Invoke a whitelisted editor menu action",
	Long: `Invoke a whitelisted editor menu action.

Available actions:
  scene/save        Save the currently edited scene
  scene/save_all    Save all open scenes
  scene/reload      Reload current scene from disk
  filesystem/scan   Trigger a full asset filesystem scan`,
	Args: cobra.ExactArgs(1),
	RunE: runMenu,
}

func init() {
	rootCmd.AddCommand(menuCmd)
}

func runMenu(_ *cobra.Command, args []string) error {
	action := args[0]
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

	resp, err := cl.Send(ctx, "menu", map[string]any{"action": action})
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
	fmt.Printf("Action invoked: %v\n", data["action"])
	return nil
}
