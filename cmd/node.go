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
	flagNodeDepth int
	flagNodeProps bool
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Inspect and modify the edited scene tree",
}

var nodeTreeCmd = &cobra.Command{
	Use:   "tree [path]",
	Short: "Print the scene tree starting at path (default: scene root)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runNodeTree,
}
var nodeGetCmd = &cobra.Command{
	Use:   "get <path> <property>",
	Short: "Read a node property",
	Args:  cobra.ExactArgs(2),
	RunE:  runNodeGet,
}
var nodeSetCmd = &cobra.Command{
	Use:   "set <path> <property> <value>",
	Short: "Write a node property (requires enable_exec)",
	Args:  cobra.ExactArgs(3),
	RunE:  runNodeSet,
}

func init() {
	rootCmd.AddCommand(nodeCmd)
	nodeCmd.AddCommand(nodeTreeCmd)
	nodeCmd.AddCommand(nodeGetCmd)
	nodeCmd.AddCommand(nodeSetCmd)
	nodeTreeCmd.Flags().IntVar(&flagNodeDepth, "depth", 3, "tree traversal depth")
	nodeTreeCmd.Flags().BoolVar(&flagNodeProps, "props", false, "include node properties in tree output")
}

func nodeCall(params map[string]any) {
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

	resp, err := cl.Send(ctx, "node", params)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connection error:", err)
		os.Exit(clierrors.ExitNoConnection)
	}
	if !resp.Success {
		code := ""
		if resp.Error != nil {
			code = resp.Error.Code
		}
		switch code {
		case "UNAUTHORIZED":
			fmt.Fprintln(os.Stderr, "authentication failed")
			os.Exit(clierrors.ExitUnauthorized)
		case "EXEC_DISABLED":
			fmt.Fprintln(os.Stderr, "node set is disabled; enable via Editor > Editor Settings > godot_cli/enable_exec")
			os.Exit(clierrors.ExitCommandError)
		default:
			if resp.Error != nil {
				fmt.Fprintln(os.Stderr, "error:", resp.Error.Message)
			}
			os.Exit(clierrors.ExitCommandError)
		}
	}
	if flagJSON {
		output.PrintJSON(resp.Data)
		return
	}
	var generic map[string]any
	_ = json.Unmarshal(resp.Data, &generic)
	out, _ := json.MarshalIndent(generic, "", "  ")
	fmt.Println(string(out))
}

func runNodeTree(_ *cobra.Command, args []string) error {
	path := ""
	if len(args) > 0 {
		path = args[0]
	}
	nodeCall(map[string]any{
		"action": "tree",
		"path":   path,
		"depth":  flagNodeDepth,
		"props":  flagNodeProps,
	})
	return nil
}

func runNodeGet(_ *cobra.Command, args []string) error {
	nodeCall(map[string]any{
		"action":   "get",
		"path":     args[0],
		"property": args[1],
	})
	return nil
}

func runNodeSet(_ *cobra.Command, args []string) error {
	var parsed any
	if err := json.Unmarshal([]byte(args[2]), &parsed); err != nil {
		parsed = args[2]
	}
	nodeCall(map[string]any{
		"action":   "set",
		"path":     args[0],
		"property": args[1],
		"value":    parsed,
	})
	return nil
}
