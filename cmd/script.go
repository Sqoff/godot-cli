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

var flagScriptMethod string

var scriptCmd = &cobra.Command{
	Use:   "script <file>",
	Short: "Run a GDScript file inside the editor (requires enable_exec)",
	Args:  cobra.ExactArgs(1),
	RunE:  runScript,
}

func init() {
	rootCmd.AddCommand(scriptCmd)
	scriptCmd.Flags().StringVar(&flagScriptMethod, "method", "_cli_execute", "method to call after attaching the script")
}

func runScript(_ *cobra.Command, args []string) error {
	data, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot read script file:", err)
		os.Exit(clierrors.ExitCommandError)
	}

	inst, err := cmdutil.Resolve(flagPort, flagProject)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(clierrors.ExitNoConnection)
	}
	tr := client.NewHTTPTransport("127.0.0.1", inst.Port, inst.Token)
	defer tr.Close()
	cl := client.New(tr)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := cl.Send(ctx, "script", map[string]any{
		"code":   string(data),
		"method": flagScriptMethod,
	})
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
			fmt.Fprintln(os.Stderr, "script is disabled; enable via Editor > Editor Settings > godot_cli/enable_exec")
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
	var generic map[string]any
	if err := json.Unmarshal(resp.Data, &generic); err != nil {
		output.PrintJSON(resp.Data)
		return nil
	}
	out, _ := json.MarshalIndent(generic["result"], "", "  ")
	fmt.Println(string(out))
	return nil
}
