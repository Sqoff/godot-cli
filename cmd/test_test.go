package cmd

import "testing"

func TestTestCmdRegistered(t *testing.T) {
	if !hasCommand("test") {
		t.Fatal("test command not registered")
	}
}

func TestTestFlags(t *testing.T) {
	for _, name := range []string{"directory", "filter", "mode", "godot"} {
		if testCmd.Flags().Lookup(name) == nil {
			t.Errorf("missing --%s flag", name)
		}
	}
}
