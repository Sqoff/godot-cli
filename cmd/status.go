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

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show connected editor status",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(_ *cobra.Command, _ []string) error {
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

	resp, err := cl.Send(ctx, "status", nil)
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
	fmt.Printf("Project:       %v\n", data["project"])
	fmt.Printf("Godot version: %v\n", data["godot_version"])
	fmt.Printf("Scene:         %v\n", data["current_scene"])
	return nil
}
