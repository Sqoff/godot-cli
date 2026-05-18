package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/Sqoff/godot-cli/internal/client"
	"github.com/Sqoff/godot-cli/internal/cmdutil"
	clierrors "github.com/Sqoff/godot-cli/internal/errors"
	"github.com/Sqoff/godot-cli/internal/output"
)

var (
	flagTestDir    string
	flagTestFilter string
	flagTestMode   string
	flagTestGodot  string
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run GUT (Godot Unit Test) tests in this project",
	RunE:  runTest,
}

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.Flags().StringVar(&flagTestDir, "directory", "", "test directory (default: res://test or res://tests)")
	testCmd.Flags().StringVar(&flagTestFilter, "filter", "", "filter pattern for test names")
	testCmd.Flags().StringVar(&flagTestMode, "mode", "standalone", "execution mode: standalone|editor")
	testCmd.Flags().StringVar(&flagTestGodot, "godot", "", "godot binary path (default: 'godot' or 'godot4' on PATH)")
}

func runTest(_ *cobra.Command, _ []string) error {
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

	resp, err := cl.Send(ctx, "test", map[string]any{
		"directory": flagTestDir,
		"filter":    flagTestFilter,
		"mode":      flagTestMode,
	})
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
	var data struct {
		Installed bool   `json:"installed"`
		Mode      string `json:"mode"`
		Directory string `json:"directory"`
		Command   string `json:"command"`
		Message   string `json:"message"`
		Advice    string `json:"advice"`
	}
	_ = json.Unmarshal(resp.Data, &data)

	if !data.Installed {
		fmt.Fprintln(os.Stderr, "GUT addon not installed at res://addons/gut. Install via AssetLib or https://github.com/bitwes/Gut")
		if flagJSON {
			output.PrintJSON(resp.Data)
		}
		os.Exit(clierrors.ExitCommandError)
	}

	if flagJSON {
		output.PrintJSON(resp.Data)
		return nil
	}

	if flagTestMode != "standalone" {
		if data.Message != "" {
			fmt.Println(data.Message)
		}
		if data.Advice != "" {
			fmt.Println(data.Advice)
		}
		return nil
	}

	godotBin := flagTestGodot
	if godotBin == "" {
		godotBin = locateGodot()
	}
	if godotBin == "" {
		fmt.Println("GUT installed.")
		fmt.Println("Run the suggested command yourself, or rerun with --godot <path>:")
		fmt.Println("  " + data.Command)
		return nil
	}

	dir := data.Directory
	if dir == "" {
		dir = "res://test/"
	}
	args := []string{"--headless", "-s", "addons/gut/gut_cmdln.gd", "-gdir=" + dir, "-gexit"}
	if flagTestFilter != "" {
		args = append(args, "-gtest="+flagTestFilter)
	}
	c := exec.Command(godotBin, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			os.Exit(ee.ExitCode())
		}
		fmt.Fprintln(os.Stderr, "test failed:", err)
		os.Exit(clierrors.ExitCommandError)
	}
	return nil
}
