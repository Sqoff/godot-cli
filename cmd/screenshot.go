package cmd

import (
	"context"
	"encoding/base64"
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
	flagScreenshotOutput string
	flagScreenshotTarget string
)

var screenshotCmd = &cobra.Command{
	Use:   "screenshot",
	Short: "Capture editor or game viewport as PNG",
	RunE:  runScreenshot,
}

func init() {
	rootCmd.AddCommand(screenshotCmd)
	screenshotCmd.Flags().StringVar(&flagScreenshotOutput, "output", "", "output PNG file path (required)")
	screenshotCmd.Flags().StringVar(&flagScreenshotTarget, "target", "game", "capture target: editor|game")
}

func runScreenshot(_ *cobra.Command, _ []string) error {
	if flagScreenshotOutput == "" {
		fmt.Fprintln(os.Stderr, "error: --output is required")
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := cl.Send(ctx, "screenshot", map[string]any{"target": flagScreenshotTarget})
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
	var data struct {
		PngBase64 string `json:"png_base64"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		fmt.Fprintln(os.Stderr, "decode failed:", err)
		os.Exit(clierrors.ExitCommandError)
	}
	raw, err := base64.StdEncoding.DecodeString(data.PngBase64)
	if err != nil {
		fmt.Fprintln(os.Stderr, "base64 decode failed:", err)
		os.Exit(clierrors.ExitCommandError)
	}
	if err := os.WriteFile(flagScreenshotOutput, raw, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "write failed:", err)
		os.Exit(clierrors.ExitCommandError)
	}
	fmt.Printf("Saved %dx%d PNG to %s\n", data.Width, data.Height, flagScreenshotOutput)
	return nil
}
