package cmd

import "testing"

func TestWatchCmdRegistered(t *testing.T) {
	if !hasCommand("watch") {
		t.Fatal("watch command not registered")
	}
}

func TestWatchFlags(t *testing.T) {
	for _, name := range []string{"path", "interval", "once"} {
		if watchCmd.Flags().Lookup(name) == nil {
			t.Errorf("missing --%s flag", name)
		}
	}
}
