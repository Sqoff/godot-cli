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
	flagSettingPrefix string
	flagSettingAll    bool
)

var settingCmd = &cobra.Command{
	Use:   "setting",
	Short: "Read and write ProjectSettings entries",
}

var settingGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Read a setting value",
	Args:  cobra.ExactArgs(1),
	RunE:  runSettingGet,
}
var settingSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Write a setting and save project.godot (requires enable_exec)",
	Args:  cobra.ExactArgs(2),
	RunE:  runSettingSet,
}
var settingListCmd = &cobra.Command{
	Use:   "list",
	Short: "List settings (user-defined by default; use --all for built-ins)",
	RunE:  runSettingList,
}

func init() {
	rootCmd.AddCommand(settingCmd)
	settingCmd.AddCommand(settingGetCmd)
	settingCmd.AddCommand(settingSetCmd)
	settingCmd.AddCommand(settingListCmd)
	settingListCmd.Flags().StringVar(&flagSettingPrefix, "prefix", "", "limit to settings starting with prefix")
	settingListCmd.Flags().BoolVar(&flagSettingAll, "all", false, "include built-in default categories")
}

func settingCall(params map[string]any) {
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

	resp, err := cl.Send(ctx, "setting", params)
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
			fmt.Fprintln(os.Stderr, "setting set is disabled; enable via Editor > Editor Settings > godot_cli/enable_exec")
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

func runSettingGet(_ *cobra.Command, args []string) error {
	settingCall(map[string]any{"action": "get", "key": args[0]})
	return nil
}

func runSettingSet(_ *cobra.Command, args []string) error {
	var parsed any
	if err := json.Unmarshal([]byte(args[1]), &parsed); err != nil {
		parsed = args[1]
	}
	settingCall(map[string]any{"action": "set", "key": args[0], "value": parsed})
	return nil
}

func runSettingList(_ *cobra.Command, _ []string) error {
	settingCall(map[string]any{
		"action": "list",
		"prefix": flagSettingPrefix,
		"all":    flagSettingAll,
	})
	return nil
}
