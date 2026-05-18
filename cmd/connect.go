package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/Sqoff/godot-cli/internal/client"
	"github.com/Sqoff/godot-cli/internal/cmdutil"
	clierrors "github.com/Sqoff/godot-cli/internal/errors"
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Verify editor connection",
	RunE:  runConnect,
}

func init() {
	rootCmd.AddCommand(connectCmd)
}

func runConnect(_ *cobra.Command, _ []string) error {
	inst, err := cmdutil.Resolve(flagPort, flagProject)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(clierrors.ExitNoConnection)
	}

	tr := client.NewHTTPTransport("127.0.0.1", inst.Port, inst.Token)
	defer tr.Close()
	cl := client.New(tr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cl.Ping(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "connection failed:", err)
		os.Exit(clierrors.ExitNoConnection)
	}

	fmt.Printf("Connected  port=%d  project=%s  godot=%s\n",
		inst.Port, inst.ProjectPath, inst.GodotVersion)
	return nil
}
