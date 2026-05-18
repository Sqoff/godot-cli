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

var sceneCmd = &cobra.Command{
	Use:   "scene",
	Short: "Get the current editor scene",
	RunE:  runSceneGet,
}

var sceneSetCmd = &cobra.Command{
	Use:   "set <path>",
	Short: "Open a scene in the editor",
	Args:  cobra.ExactArgs(1),
	RunE:  runSceneSet,
}

func init() {
	rootCmd.AddCommand(sceneCmd)
	sceneCmd.AddCommand(sceneSetCmd)
}

func runSceneGet(_ *cobra.Command, _ []string) error {
	return doSceneRequest("get", "")
}

func runSceneSet(_ *cobra.Command, args []string) error {
	return doSceneRequest("set", args[0])
}

func doSceneRequest(action, path string) error {
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
	if path != "" {
		params["path"] = path
	}

	resp, err := cl.Send(ctx, "scene", params)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connection error:", err)
		os.Exit(clierrors.ExitNoConnection)
	}

	if !resp.Success {
		code := ""
		if resp.Error != nil {
			code = resp.Error.Code
		}
		if code == "UNAUTHORIZED" {
			fmt.Fprintln(os.Stderr, "authentication failed")
			os.Exit(clierrors.ExitUnauthorized)
		}
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
	fmt.Printf("Scene: %v\n", data["scene"])
	return nil
}
