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

var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Search and inspect project resources",
}

var resourceFindCmd = &cobra.Command{
	Use:   "find <pattern>",
	Short: "Find resources matching a path substring",
	Args:  cobra.ExactArgs(1),
	RunE:  runResourceFind,
}

var resourceInfoCmd = &cobra.Command{
	Use:   "info <res://path>",
	Short: "Show information about a single resource",
	Args:  cobra.ExactArgs(1),
	RunE:  runResourceInfo,
}

func init() {
	rootCmd.AddCommand(resourceCmd)
	resourceCmd.AddCommand(resourceFindCmd)
	resourceCmd.AddCommand(resourceInfoCmd)
}

func resourceCall(action string, params map[string]any) {
	inst, err := cmdutil.Resolve(flagPort, flagProject)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(clierrors.ExitNoConnection)
	}
	tr := client.NewHTTPTransport("127.0.0.1", inst.Port, inst.Token)
	defer tr.Close()
	cl := client.New(tr)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	params["action"] = action
	resp, err := cl.Send(ctx, "resource", params)
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
		return
	}
	var generic map[string]any
	_ = json.Unmarshal(resp.Data, &generic)
	switch action {
	case "find":
		paths, _ := generic["paths"].([]any)
		for _, p := range paths {
			fmt.Println(p)
		}
		fmt.Printf("(%d match)\n", len(paths))
	case "info":
		out, _ := json.MarshalIndent(generic, "", "  ")
		fmt.Println(string(out))
	}
}

func runResourceFind(_ *cobra.Command, args []string) error {
	resourceCall("find", map[string]any{"pattern": args[0]})
	return nil
}

func runResourceInfo(_ *cobra.Command, args []string) error {
	resourceCall("info", map[string]any{"path": args[0]})
	return nil
}
