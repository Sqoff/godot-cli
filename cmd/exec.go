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

var execCmd = &cobra.Command{
	Use:   "exec <node_path> <method> [args...]",
	Short: "Call a method on a scene node (requires enable_exec)",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runExec,
}

func init() {
	rootCmd.AddCommand(execCmd)
}

func runExec(_ *cobra.Command, args []string) error {
	nodePath := args[0]
	method := args[1]
	methodArgs := make([]any, len(args)-2)
	for i, a := range args[2:] {
		methodArgs[i] = a
	}

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

	params := map[string]any{
		"node":   nodePath,
		"method": method,
		"args":   methodArgs,
	}

	resp, err := cl.Send(ctx, "exec", params)
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
			fmt.Fprintln(os.Stderr, "exec is disabled; enable via Editor > Editor Settings > godot_cli/enable_exec")
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
		return nil
	}

	var data map[string]any
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		output.PrintJSON(resp.Data)
		return nil
	}
	result := data["result"]
	if result == nil {
		fmt.Println("(null)")
		return nil
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(out))
	return nil
}
