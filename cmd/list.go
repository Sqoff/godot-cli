package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"github.com/Sqoff/godot-cli/internal/client"
	"github.com/Sqoff/godot-cli/internal/cmdutil"
	"github.com/Sqoff/godot-cli/internal/output"
)

var flagListLocal bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available commands (queries editor unless --local)",
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVar(&flagListLocal, "local", false, "list CLI-side commands without querying editor")
}

// localCommands enumerates every command the CLI knows about.
// This is shown when --local is set or when no editor is reachable.
var localCommands = []string{
	"connect",
	"exec",
	"export",
	"init",
	"list",
	"log",
	"menu",
	"node",
	"pause",
	"plugin",
	"profiler",
	"refresh",
	"reserialize",
	"resource",
	"run",
	"scene",
	"screenshot",
	"script",
	"setting",
	"status",
	"stop",
	"test",
	"update",
	"watch",
}

func runList(_ *cobra.Command, _ []string) error {
	if flagListLocal {
		printCommandList(localCommands, "Available commands (CLI-side)")
		return nil
	}

	inst, err := cmdutil.Resolve(flagPort, flagProject)
	if err != nil {
		fmt.Fprintln(os.Stderr, "no editor connection; showing CLI-side list (use --local to suppress this fallback)")
		printCommandList(localCommands, "Available commands (CLI-side)")
		return nil
	}

	tr := client.NewHTTPTransport("127.0.0.1", inst.Port, inst.Token)
	defer tr.Close()
	cl := client.New(tr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := cl.Send(ctx, "list", nil)
	if err != nil || !resp.Success {
		fmt.Fprintln(os.Stderr, "editor list query failed; showing CLI-side list")
		printCommandList(localCommands, "Available commands (CLI-side)")
		return nil
	}

	if flagJSON {
		output.PrintJSON(resp.Data)
		return nil
	}

	var data struct {
		Commands []string `json:"commands"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil || len(data.Commands) == 0 {
		printCommandList(localCommands, "Available commands (CLI-side)")
		return nil
	}
	printCommandList(data.Commands, "Available commands (editor-registered)")
	return nil
}

func printCommandList(commands []string, header string) {
	sorted := make([]string, len(commands))
	copy(sorted, commands)
	sort.Strings(sorted)

	fmt.Println(header + ":")
	for _, c := range sorted {
		fmt.Printf("  %s\n", c)
	}
}

