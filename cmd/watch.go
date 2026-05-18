package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/Sqoff/godot-cli/internal/client"
	"github.com/Sqoff/godot-cli/internal/cmdutil"
	clierrors "github.com/Sqoff/godot-cli/internal/errors"
)

var (
	flagWatchPath     string
	flagWatchInterval int
	flagWatchOnce     bool
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch project files and trigger filesystem rescans on change",
	RunE:  runWatch,
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().StringVar(&flagWatchPath, "path", "res://", "root path to watch")
	watchCmd.Flags().IntVar(&flagWatchInterval, "interval", 2, "poll interval seconds")
	watchCmd.Flags().BoolVar(&flagWatchOnce, "once", false, "single snapshot then exit")
}

type watchFile struct {
	Path  string `json:"path"`
	Mtime int64  `json:"mtime"`
}

type watchSnapshot struct {
	Files []watchFile `json:"files"`
}

func runWatch(_ *cobra.Command, _ []string) error {
	inst, err := cmdutil.Resolve(flagPort, flagProject)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(clierrors.ExitNoConnection)
	}
	tr := client.NewHTTPTransport("127.0.0.1", inst.Port, inst.Token)
	defer tr.Close()
	cl := client.New(tr)

	snap, err := fetchWatchSnapshot(cl)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(clierrors.ExitNoConnection)
	}

	prev := indexSnapshot(snap)
	if flagWatchOnce {
		paths := make([]string, 0, len(prev))
		for p := range prev {
			paths = append(paths, p)
		}
		sort.Strings(paths)
		for _, p := range paths {
			fmt.Println(p)
		}
		fmt.Printf("(%d files)\n", len(prev))
		return nil
	}

	fmt.Printf("Watching %s every %ds (Ctrl-C to stop)\n", flagWatchPath, flagWatchInterval)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	ticker := time.NewTicker(time.Duration(flagWatchInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigCh:
			fmt.Println("\nstopped")
			return nil
		case <-ticker.C:
			snap, err := fetchWatchSnapshot(cl)
			if err != nil {
				fmt.Fprintln(os.Stderr, "warn:", err)
				continue
			}
			cur := indexSnapshot(snap)
			changes := diffSnapshots(prev, cur)
			if len(changes) > 0 {
				for _, line := range changes {
					fmt.Println(line)
				}
				if _, err := sendWatchAction(cl, "scan"); err != nil {
					fmt.Fprintln(os.Stderr, "warn: scan failed:", err)
				}
			}
			prev = cur
		}
	}
}

func fetchWatchSnapshot(cl *client.Client) (*watchSnapshot, error) {
	raw, err := sendWatchAction(cl, "snapshot")
	if err != nil {
		return nil, err
	}
	var s watchSnapshot
	_ = json.Unmarshal(raw, &s)
	return &s, nil
}

func sendWatchAction(cl *client.Client, action string) (json.RawMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	resp, err := cl.Send(ctx, "watch", map[string]any{"action": action, "path": flagWatchPath})
	if err != nil {
		return nil, fmt.Errorf("connection error: %w", err)
	}
	if !resp.Success {
		if resp.Error != nil {
			return nil, fmt.Errorf("%s", resp.Error.Message)
		}
		return nil, fmt.Errorf("watch %s failed", action)
	}
	return resp.Data, nil
}

func indexSnapshot(s *watchSnapshot) map[string]int64 {
	out := make(map[string]int64, len(s.Files))
	for _, f := range s.Files {
		out[f.Path] = f.Mtime
	}
	return out
}

func diffSnapshots(prev, cur map[string]int64) []string {
	var out []string
	for p, m := range cur {
		if pm, ok := prev[p]; !ok {
			out = append(out, "[added]    "+p)
		} else if pm != m {
			out = append(out, "[modified] "+p)
		}
	}
	for p := range prev {
		if _, ok := cur[p]; !ok {
			out = append(out, "[removed]  "+p)
		}
	}
	sort.Strings(out)
	return out
}
